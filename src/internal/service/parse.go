
package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ReqGet struct {
	ID               int
	Depth            int
	IncludeEmployees bool
}

type ReqPostDepartment struct {
	Name     string
	ParentID *int
}

type ReqPostEmployee struct {
	DepartmentID int
	FullName     string
	Position     string
	HiredAt      *time.Time
}

type ReqPatch struct {
	ID        int
	Name      *string
	SetParent bool
	ParentID  *int
}

type ReqDelete struct {
	ID         int
	Mode       string
	ReassignTo *int
}


func ParseGetRequest(r *http.Request, id int) (*ReqGet, error) {
	if r.Method != http.MethodGet {
		return nil, errors.New("method not allowed")
	}
	if id <= 0 {
		return nil, errors.New("invalid department id")
	}
	return &ReqGet{
		ID:               id,
		Depth:            getQueryInt(r, "depth", 1, 1, 5),
		IncludeEmployees: getQueryBool(r, "include_employees", true),
	}, nil
}

func ParsePostDepartment(r *http.Request) (*ReqPostDepartment, error) {
	if r.Method != http.MethodPost {
		return nil, errors.New("method not allowed")
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, errors.New("invalid JSON body")
	}
	req := &ReqPostDepartment{}
	if name, ok := body["name"].(string); ok {
		req.Name = strings.TrimSpace(name)
	}
	if parentIDRaw, ok := body["parent_id"]; ok && parentIDRaw != nil {
		switch v := parentIDRaw.(type) {
		case float64:
			val := int(v)
			req.ParentID = &val
		case int:
			req.ParentID = &v
		}
	}
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if len(req.Name) > 200 {
		return nil, errors.New("name too long (max 200)")
	}
	return req, nil
}

func ParsePostEmployee(r *http.Request, departmentID int) (*ReqPostEmployee, error) {
	if r.Method != http.MethodPost {
		return nil, errors.New("method not allowed")
	}
	if departmentID <= 0 {
		return nil, errors.New("invalid department id")
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, errors.New("invalid JSON body")
	}
	req := &ReqPostEmployee{DepartmentID: departmentID}
	if fullName, ok := body["full_name"].(string); ok {
		req.FullName = strings.TrimSpace(fullName)
	}
	if position, ok := body["position"].(string); ok {
		req.Position = strings.TrimSpace(position)
	}
	if hiredAtStr, ok := body["hired_at"].(string); ok && hiredAtStr != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(hiredAtStr))
		if err != nil {
			return nil, errors.New("hired_at must be YYYY-MM-DD")
		}
		req.HiredAt = &t
	}
	if req.FullName == "" {
		return nil, errors.New("full_name is required")
	}
	if len(req.FullName) > 200 {
		return nil, errors.New("full_name too long (max 200)")
	}
	if req.Position == "" {
		return nil, errors.New("position is required")
	}
	if len(req.Position) > 200 {
		return nil, errors.New("position too long (max 200)")
	}
	return req, nil
}

func ParsePatchRequest(r *http.Request, id int) (*ReqPatch, error) {
	if r.Method != http.MethodPatch {
		return nil, errors.New("method not allowed")
	}
	if id <= 0 {
		return nil, errors.New("invalid department id")
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, errors.New("invalid JSON body")
	}
	req := &ReqPatch{ID: id}
	if nameRaw, ok := body["name"]; ok {
		name, ok := nameRaw.(string)
		if !ok {
			return nil, errors.New("name must be a string")
		}
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return nil, errors.New("name cannot be empty")
		}
		if len(trimmed) > 200 {
			return nil, errors.New("name too long (max 200)")
		}
		req.Name = &trimmed
	}
	if _, ok := body["parent_id"]; ok {
		req.SetParent = true
		if parentIDRaw := body["parent_id"]; parentIDRaw != nil {
			switch v := parentIDRaw.(type) {
			case float64:
				val := int(v)
				req.ParentID = &val
			case int:
				req.ParentID = &v
			default:
				return nil, errors.New("invalid parent_id type")
			}
		}
		// nil → явный перевод в корень
	}
	return req, nil
}

func ParseDeleteRequest(r *http.Request, id int) (*ReqDelete, error) {
	if r.Method != http.MethodDelete {
		return nil, errors.New("method not allowed")
	}
	if id <= 0 {
		return nil, errors.New("invalid department id")
	}
	req := &ReqDelete{ID: id, Mode: r.URL.Query().Get("mode")}

	if req.Mode != "cascade" && req.Mode != "reassign" {
		return nil, errors.New("mode must be 'cascade' or 'reassign'")
	}

	if req.Mode == "reassign" {
		rt := r.URL.Query().Get("reassign_to_department_id")
		if rt == "" {
			return nil, errors.New("reassign_to_department_id required for mode=reassign")
		}
		val, err := strconv.Atoi(rt)
		if err != nil || val <= 0 {
			return nil, errors.New("invalid reassign_to_department_id")
		}
		req.ReassignTo = &val
	}
	return req, nil
}
