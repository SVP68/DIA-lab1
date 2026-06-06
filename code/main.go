package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"os/exec"
)

// SolarEquipment - Первая модель-коллекция (Услуги)
type SolarEquipment struct {
	ID          	int
	ModelName   	string
	Price       	float64
	CapacityPower	int
	Description 	string
	ImageKey    	string // Ключ для Minio (латиница)
	VideoKey    	string // Ключ для Minio (латиница)
}

// CapacityPowerStr - Вторая модель-коллекция (Состав заявки)
type CapacityPowerStr struct {
	RequestID   	int
	EquipmentID 	int
	Quantity    	int    // Поле м-м
}

// Имитация базы данных (In-Memory коллекции)
var equipments = []SolarEquipment{
	{1, "Аккумулятор Vektor GL 12-100", 10000.0, 100, "Аккумуляторные батареи VEKTOR ENERGY cерии GEL (GL) изготовлены по технологии AGM+GEL. Электролит в данных аккумуляторах увязан в гель по средством оксида кремния SiO2, но также как в стандартных аккумуляторах, используется AGM сепаратор. Аккумуляторы серии GL имеют отличные разрядные и эксплуатационные характеристики.", "AKB100.jpg", "solar.webm"},
	{2, "Аккумулятор Vektor GL 12-150", 15000.0, 150, "Аккумуляторные батареи VEKTOR ENERGY cерии GEL (GL) изготовлены по технологии AGM+GEL. Электролит в данных аккумуляторах увязан в гель по средством оксида кремния SiO2, но также как в стандартных аккумуляторах, используется AGM сепаратор. Аккумуляторы серии GL имеют отличные разрядные и эксплуатационные характеристики.", "AKB150.jpg", "solar.webm"},
	{3, "Аккумулятор Vektor GL 12-200", 20000.0, 200, "Аккумуляторные батареи VEKTOR ENERGY cерии GEL (GL) изготовлены по технологии AGM+GEL. Электролит в данных аккумуляторах увязан в гель по средством оксида кремния SiO2, но также как в стандартных аккумуляторах, используется AGM сепаратор. Аккумуляторы серии GL имеют отличные разрядные и эксплуатационные характеристики.", "AKB200.jpg", "solar.webm"},
	{4, "Солнечная панель DELTA_NXT500", 12000.0, 500, "Фотоэлектрический солнечный модуль (ФСМ) DELTA NXT 400-54/2 M10 HC DELTA NXT - это серия фотоэлектрических модулей, выполненных из материалов экстра-класса. При невысокой интенсивности солнечного излучения, DELTA NXT вырабатывают больше электроэнергии, чем стандартные солнечные модули с аналогичными характеристиками.", "DELTA_NXT500.jpg", "solar1.webm"},
	{5, "Солнечная панель GWS280", 89000.0, 280, "При производстве солнечных панелей GWS используются высококачественные материалы, что гарантирует наивысшее качество изделий: прочный защитный слой специального закалённого стекла и усиленная рамка из анодированного алюминия, устойчивая к коррозии, обеспечивает высокий класс защиты от механических повреждений, влаги и высокое сопротивление экстремальной ветровой нагрузке.", "GWS280.jpg", "solar1.webm"},
	{6, "Солнечная панель M300WT", 26000.0, 300, "Солнечные батареи серии М300 являются фотоэлектрическими модулями, выполненными из материалов экстра-класса. При невысокой интенсивности солнечного излучения, вырабатывают больше электроэнергии, чем стандартные солнечные модули с аналогичными характеристиками.", "M300WT.jpg", "solar1.webm"},
}

// Словарь заявок (ключ - ID связи, для простоты используем срез)
var currentCapacityPower = []CapacityPowerStr{
	{RequestID: 101, EquipmentID: 4, Quantity: 4},
	{RequestID: 101, EquipmentID: 2, Quantity: 1},
	{RequestID: 101, EquipmentID: 6, Quantity: 2},
}

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

func main() {
		
	containerName := "minio-go-launcher"
	dataPath := "./minio_storage" 
	
	// 1. Команда для запуска MinIO через WSL в Docker
	// -d: запуск в фоновом режиме
	// --rm: автоматически удаляется старый контейнер
	// -p: проброс портов (9000 для API, 9001 для консоли)
	cmdArgs := []string{
		"-d", "Ubuntu", "docker", "run", "-d",
		"--name", containerName,
		"-p", "9000:9000",
		"-p", "9001:9001",
		"-v", dataPath+":/data", //данные сохраняются в dataPath
		"-e", "MINIO_ROOT_USER=admin",
		"-e", "MINIO_ROOT_PASSWORD=adminPass",
		"quay.io/minio/minio", "server", "/data", "--console-address", ":9001",
	}

	// Сначала пробуем принудительно удалить старый контейнер
	exec.Command("docker", "rm", "-f", containerName).Run()

	cmd := exec.Command("wsl", cmdArgs...)
	
	// сделаем доступ для папки solar Custom, чтобы были видны картинки из Minio
	 exec.Command("docker", "-it", containerName, "/bin/sh  mc alias set local http://localhost:9000 admin adminPass && mc anonymous set download local/solar").Run()

	// Выполнение команды
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка при запуске: %v\nКонсоль: %s\n", err, string(output))
		return
	}

	log.Println("Сервер MinIO успешно запущен на http://localhost:9001")
	
	// Статические файлы (CSS)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 3 GET запроса
	http.HandleFunc("/", handleCatalog)          // 1. Список услуг
	http.HandleFunc("/equipment/", handleDetail) // 2. Подробно (Vibes)
	http.HandleFunc("/request/", handleCart)     // 3. Состав заявки

	log.Println("Сервер SolarCalc успешно запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// 1. Обработчик списка услуг
func handleCatalog(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	var filtered []SolarEquipment

	for _, e := range equipments {
		if search == "" || strings.Contains(strings.ToLower(e.ModelName), strings.ToLower(search)) {
			filtered = append(filtered, e)
		}
	}

	// Вычисляем общее количество услуг в заявке (CartCount)
	totalCount := len(currentCapacityPower)

	tmpl.ExecuteTemplate(w, "solar_index.html", map[string]interface{}{
		"Items":     filtered,
		"CartCount": totalCount,
		"Search":    search,
	})
}

// 2. Обработчик одной услуги (Vibes)
func handleDetail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/equipment/"))
	var found *SolarEquipment

	for _, e := range equipments {
		if e.ID == id {
			found = &e
			break
		}
	}

	tmpl.ExecuteTemplate(w, "solar_item.html", map[string]interface{}{
		"Equipment": found,
	})
}

// 3. Обработчик состава заявки
func handleCart(w http.ResponseWriter, r *http.Request) {
	reqID, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/request/"))

	type CartViewItem struct {
		SolarEquipment
		CapacityPowerStr
	}
	var viewItems []CartViewItem

	for _, ri := range currentCapacityPower {
		if ri.RequestID == reqID {
			for _, eq := range equipments {
				if eq.ID == ri.EquipmentID {
					viewItems = append(viewItems, CartViewItem{eq, ri})
				}
			}
		}
	}

	// Вычисляем общее количество услуг в заявке (CartCount)
	totalCount := len(currentCapacityPower)

	tmpl.ExecuteTemplate(w, "solar_request.html", map[string]interface{}{
		"RequestID": reqID,
		"CartCount": totalCount,
		"CartItems": viewItems,
	})
}
