package handlers

import (
	"errors"
	"net/http"
	"strings"

	"main/database/crud"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case r.Method == http.MethodPost && len(parts) == 1 && parts[0] == "departments":
		h.postDepartment(w, r)
	case r.Method == http.MethodPost && len(parts) == 3 && parts[0] == "departments" && parts[2] == "employees":
		h.postEmployee(w, r)
	case r.Method == http.MethodGet && len(parts) == 2 && parts[0] == "departments":
		h.getDepartment(w, r)
	case r.Method == http.MethodPatch && len(parts) == 2 && parts[0] == "departments":
		h.patchDepartment(w, r)
	case r.Method == http.MethodDelete && len(parts) == 2 && parts[0] == "departments":
		h.deleteDepartment(w, r)
	default:
		RespondError(w, http.StatusNotFound, "not found")
	}
}

func (h *Handler) postDepartment(w http.ResponseWriter, r *http.Request) {
	req, err := ParsePostDepartment(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	dept, err := crud.CreateDepartment(r.Context(), h.DB, req.Name, req.ParentID)
	if err != nil {
		switch {
		case errors.Is(err, crud.ErrNotFound):
			RespondError(w, http.StatusNotFound, "parent department not found")
		case errors.Is(err, crud.ErrConflict):
			RespondError(w, http.StatusConflict, "department name already exists in this parent")
		default:
			RespondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondJSON(w, http.StatusCreated, dept)
}

func (h *Handler) postEmployee(w http.ResponseWriter, r *http.Request) {
	req, err := ParsePostEmployee(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	emp, err := crud.AddEmployee(r.Context(), h.DB, req.DepartmentID, req.FullName, req.Position, req.HiredAt)
	if err != nil {
		if errors.Is(err, crud.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "department not found")
		} else {
			RespondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondJSON(w, http.StatusCreated, emp)
}

func (h *Handler) getDepartment(w http.ResponseWriter, r *http.Request) {
	req, err := ParseGetRequest(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	dept, err := crud.GetDepartment(r.Context(), h.DB, req.ID, req.Depth, req.IncludeEmployees)
	if err != nil {
		if errors.Is(err, crud.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "department not found")
		} else {
			RespondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondJSON(w, http.StatusOK, dept)
}

func (h *Handler) patchDepartment(w http.ResponseWriter, r *http.Request) {
	req, err := ParsePatchRequest(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	dept, err := crud.PatchDepartment(r.Context(), h.DB, req.ID, req.Name, req.SetParent, req.ParentID)
	if err != nil {
		switch {
		case errors.Is(err, crud.ErrNotFound):
			RespondError(w, http.StatusNotFound, "department not found")
		case errors.Is(err, crud.ErrConflict):
			RespondError(w, http.StatusConflict, "department name already exists in this parent")
		case errors.Is(err, crud.ErrCycle):
			RespondError(w, http.StatusConflict, "moving department would create a cycle")
		case errors.Is(err, crud.ErrSelfParent):
			RespondError(w, http.StatusBadRequest, err.Error())
		default:
			RespondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondJSON(w, http.StatusOK, dept)
}

func (h *Handler) deleteDepartment(w http.ResponseWriter, r *http.Request) {
	req, err := ParseDeleteRequest(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = crud.DeleteDepartment(r.Context(), h.DB, req.ID, req.Mode, req.ReassignTo)
	if err != nil {
		if errors.Is(err, crud.ErrNotFound) {
			RespondError(w, http.StatusNotFound, "department not found")
		} else {
			RespondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
