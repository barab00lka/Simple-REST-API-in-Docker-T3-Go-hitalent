package main_test

import . "main/database/models"

import (
	"testing"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)
// ----- Test database setup -----
var db *gorm.DB

func TestMain(m *testing.M) {
	// Connect to your test database (adjust DSN)
	dsn := "host=localhost user=postgres_user password=t3-go dbname=postgres_db port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Ensure tables exist (optional, if schema already created via goose/migration)
	// You can skip this if the schema is already present.
	db.AutoMigrate(&Department{}, &Employee{})

	// Seed test data
	if err := seedData(); err != nil {
		panic("failed to seed data: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Clean up if needed (optional)
	// db.Exec("TRUNCATE employees, departments RESTART IDENTITY CASCADE;")

	// Exit
	os.Exit(code)
}

// ----- Seeding function -----
func seedData() error {
	// Clear existing data (to avoid duplicate conflicts)
	db.Exec("TRUNCATE employees, departments RESTART IDENTITY CASCADE;")

	// ---- Step 1: Create root departments ----
	roots := []*Department{
		{Name: "Корпоративный центр"},
		{Name: "Региональный филиал \"Запад\""},
		{Name: "Региональный филиал \"Восток\""},
	}
	for _, r := range roots {
		if err := db.Create(r).Error; err != nil {
			return err
		}
	}

	// Helper function to find a department by name (and optional parent name)
	getDeptID := func(name string, parentName *string) (int, error) {
		var d Department
		q := db.Where("name = ?", name)
		if parentName != nil {
			var parent Department
			if err := db.Where("name = ?", *parentName).First(&parent).Error; err != nil {
				return 0, err
			}
			q = q.Where("parent_id = ?", parent.ID)
		} else {
			q = q.Where("parent_id IS NULL")
		}
		err := q.First(&d).Error
		return d.ID, err
	}

	// ---- Step 2: Create first‑level child departments ----
	// Корпоративный центр -> children
	corpCenterID, _ := getDeptID("Корпоративный центр", nil)
	devDept := &Department{Name: "Департамент разработки", ParentID: &corpCenterID}
	marketingDept := &Department{Name: "Департамент маркетинга", ParentID: &corpCenterID}
	salesDept := &Department{Name: "Департамент продаж", ParentID: &corpCenterID}
	for _, d := range []*Department{devDept, marketingDept, salesDept} {
		if err := db.Create(d).Error; err != nil {
			return err
		}
	}

	// Филиал "Запад" -> children
	westID, _ := getDeptID("Региональный филиал \"Запад\"", nil)
	itWest := &Department{Name: "Отдел ИТ", ParentID: &westID}
	hrWest := &Department{Name: "Отдел кадров", ParentID: &westID}
	for _, d := range []*Department{itWest, hrWest} {
		if err := db.Create(d).Error; err != nil {
			return err
		}
	}

	// Филиал "Восток" -> child
	eastID, _ := getDeptID("Региональный филиал \"Восток\"", nil)
	itEast := &Department{Name: "Отдел ИТ", ParentID: &eastID}
	if err := db.Create(itEast).Error; err != nil {
		return err
	}

	// ---- Step 3: Second‑level child departments ----
	// Inside Департамент разработки
	devID, _ := getDeptID("Департамент разработки", &[]string{"Корпоративный центр"}[0])
	backend := &Department{Name: "Команда бэкенда", ParentID: &devID}
	frontend := &Department{Name: "Команда фронтенда", ParentID: &devID}
	devops := &Department{Name: "Команда DevOps", ParentID: &devID}
	for _, d := range []*Department{backend, frontend, devops} {
		if err := db.Create(d).Error; err != nil {
			return err
		}
	}

	// Inside Департамент маркетинга
	marketingID, _ := getDeptID("Департамент маркетинга", &[]string{"Корпоративный центр"}[0])
	content := &Department{Name: "Отдел контента", ParentID: &marketingID}
	ads := &Department{Name: "Отдел рекламы", ParentID: &marketingID}
	for _, d := range []*Department{content, ads} {
		if err := db.Create(d).Error; err != nil {
			return err
		}
	}

	// Inside Департамент продаж
	salesID, _ := getDeptID("Департамент продаж", &[]string{"Корпоративный центр"}[0])
	activeSales := &Department{Name: "Отдел активных продаж", ParentID: &salesID}
	keyAccounts := &Department{Name: "Отдел работы с ключевыми клиентами", ParentID: &salesID}
	for _, d := range []*Department{activeSales, keyAccounts} {
		if err := db.Create(d).Error; err != nil {
			return err
		}
	}

	// ---- Step 4: Create employees ----
	// Helper to parse date string into *time.Time
	parseDate := func(s string) *time.Time {
		if s == "" {
			return nil
		}
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil
		}
		return &t
	}

	// Map of department names to their IDs (for quick lookup)
	deptIDs := make(map[string]int)
	var allDepts []Department
	db.Find(&allDepts)
	for _, d := range allDepts {
		deptIDs[d.Name] = d.ID
	}

	employees := []*Employee{
		// Корпоративный центр
		{DepartmentID: deptIDs["Корпоративный центр"], FullName: "Иванов Иван Иванович", Position: "Генеральный директор", HiredAt: parseDate("2020-01-10")},
		// Департамент разработки
		{DepartmentID: deptIDs["Департамент разработки"], FullName: "Петров Пётр Петрович", Position: "Руководитель разработки", HiredAt: parseDate("2021-03-15")},
		// Команда бэкенда
		{DepartmentID: deptIDs["Команда бэкенда"], FullName: "Сидоров Алексей Владимирович", Position: "Senior Backend Engineer", HiredAt: parseDate("2022-05-20")},
		{DepartmentID: deptIDs["Команда бэкенда"], FullName: "Кузнецова Мария Игоревна", Position: "Backend Developer", HiredAt: parseDate("2023-01-10")},
		{DepartmentID: deptIDs["Команда бэкенда"], FullName: "Фёдоров Андрей Васильевич", Position: "Стажёр-разработчик", HiredAt: nil},
		// Команда фронтенда
		{DepartmentID: deptIDs["Команда фронтенда"], FullName: "Николаев Дмитрий Сергеевич", Position: "Frontend Team Lead", HiredAt: parseDate("2021-11-01")},
		// Команда DevOps
		{DepartmentID: deptIDs["Команда DevOps"], FullName: "Васильев Олег Николаевич", Position: "DevOps инженер", HiredAt: parseDate("2022-09-12")},
		// Департамент маркетинга
		{DepartmentID: deptIDs["Департамент маркетинга"], FullName: "Михайлова Анна Андреевна", Position: "Директор по маркетингу", HiredAt: parseDate("2019-07-01")},
		// Отдел контента
		{DepartmentID: deptIDs["Отдел контента"], FullName: "Егорова Екатерина Павловна", Position: "Контент-менеджер", HiredAt: parseDate("2021-02-18")},
		// Отдел рекламы
		{DepartmentID: deptIDs["Отдел рекламы"], FullName: "Смирнов Артём Денисович", Position: "PPC специалист", HiredAt: parseDate("2022-11-05")},
		// Департамент продаж
		{DepartmentID: deptIDs["Департамент продаж"], FullName: "Тихонов Максим Викторович", Position: "Руководитель отдела продаж", HiredAt: parseDate("2018-04-22")},
		// Отдел активных продаж
		{DepartmentID: deptIDs["Отдел активных продаж"], FullName: "Козлова Елена Игоревна", Position: "Менеджер по продажам", HiredAt: parseDate("2020-10-14")},
		{DepartmentID: deptIDs["Отдел активных продаж"], FullName: "Павлова Ольга Алексеевна", Position: "Ассистент отдела продаж", HiredAt: nil},
		// Отдел работы с ключевыми клиентами
		{DepartmentID: deptIDs["Отдел работы с ключевыми клиентами"], FullName: "Алексеев Владимир Сергеевич", Position: "Key Account Manager", HiredAt: parseDate("2019-12-03")},
		// Филиал "Запад"
		{DepartmentID: deptIDs["Региональный филиал \"Запад\""], FullName: "Новиков Павел Андреевич", Position: "Директор филиала", HiredAt: parseDate("2017-06-01")},
		// Отдел ИТ (Запад)
		{DepartmentID: deptIDs["Отдел ИТ"], FullName: "Морозов Дмитрий Алексеевич", Position: "Системный администратор", HiredAt: parseDate("2021-08-24")},
		// Отдел кадров (Запад)
		{DepartmentID: deptIDs["Отдел кадров"], FullName: "Григорьева Светлана Петровна", Position: "HR-менеджер", HiredAt: parseDate("2022-02-11")},
		// Филиал "Восток"
		{DepartmentID: deptIDs["Региональный филиал \"Восток\""], FullName: "Соколов Илья Владимирович", Position: "Директор филиала", HiredAt: parseDate("2019-09-17")},
		// Отдел ИТ (Восток)
		{DepartmentID: deptIDs["Отдел ИТ"], FullName: "Белова Наталья Сергеевна", Position: "IT-специалист", HiredAt: parseDate("2022-04-09")},
	}

	// Insert all employees
	for _, emp := range employees {
		if err := db.Create(emp).Error; err != nil {
			return err
		}
	}
	return nil
}

// ----- Example test that uses the seeded data -----
func TestExample(t *testing.T) {
	var count int64
	db.Model(&Department{}).Count(&count)
	t.Logf("Total departments: %d", count)

	db.Model(&Employee{}).Count(&count)
	t.Logf("Total employees: %d", count)

	// You can now run your API tests against this seeded database.
}
