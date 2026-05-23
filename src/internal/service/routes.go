package service

import (
	"net/http"
	"gorm.io/gorm"
)

// SetupRouter инициализирует маршрутизатор со всеми эндпоинтами.
func SetupRouter(db *gorm.DB) *http.ServeMux {
	h := &Handler{DB: db}
	mux := http.NewServeMux()

	// POST /departments/ – создание подразделения (без id)
	mux.HandleFunc("POST /departments/", h.postDepartment)

	// POST /departments/{id}/employees/ – создание сотрудника
	mux.HandleFunc("POST /departments/{id}/employees/", h.postEmployee)

	// GET /departments/{id}/ – получение подразделения с деревом
	mux.HandleFunc("GET /departments/{id}/", h.getDepartment)

	// PATCH /departments/{id}/ – обновление подразделения
	mux.HandleFunc("PATCH /departments/{id}/", h.patchDepartment)

	// DELETE /departments/{id}/ – удаление подразделения
	mux.HandleFunc("DELETE /departments/{id}/", h.deleteDepartment)

	return mux
}

