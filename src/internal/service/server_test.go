package service_test

// Integration tests against a live server.
// Set TEST_BASE_URL (default: http://localhost:8080).
// The server and database must already be running.
// Run: TEST_BASE_URL=http://localhost:8080 go test ./... -v

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// ---------- setup ----------

var baseURL string

func TestMain(m *testing.M) {
	baseURL = os.Getenv("TEST_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:80"
	}
	os.Exit(m.Run())
}

// ---------- client ----------

type C struct{ t *testing.T }

func client(t *testing.T) *C { return &C{t} }

func (c *C) do(method, path string, body map[string]interface{}) *http.Response {
	c.t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, baseURL+path, r)
	if err != nil {
		c.t.Fatalf("build request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("do request %s %s: %v", method, path, err)
	}
	return resp
}

// ---------- helpers ----------

func body(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	var m map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&m)
	return m
}

func mustStatus(t *testing.T, resp *http.Response, want int) map[string]interface{} {
	t.Helper()
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != want {
		t.Fatalf("want %d, got %d: %s", want, resp.StatusCode, b)
	}
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	return m
}

func intID(m map[string]interface{}) int {
	return int(m["id"].(float64))
}

// uniq avoids name collisions across test runs
func uniq(s string) string {
	return fmt.Sprintf("%s_%d", s, time.Now().UnixNano())
}

// cleanup cascades so registering the root is enough
func (c *C) cleanup(id int) {
	c.t.Cleanup(func() {
		resp := c.do(http.MethodDelete, fmt.Sprintf("/departments/%d?mode=cascade", id), nil)
		resp.Body.Close()
	})
}

// createDept is a shortcut that asserts 201 and registers cleanup
func (c *C) createDept(name string, parentID *int) map[string]interface{} {
	c.t.Helper()
	payload := map[string]interface{}{"name": name}
	if parentID != nil {
		payload["parent_id"] = *parentID
	}
	resp := c.do(http.MethodPost, "/departments/", payload)
	m := mustStatus(c.t, resp, http.StatusCreated)
	id := intID(m)
	c.cleanup(id)
	return m
}

func ptr(i int) *int { return &i }

// ============================================================
// POST /departments/
// ============================================================

func TestPostDepartment_RootCreated(t *testing.T) {
	c := client(t)
	name := uniq("Engineering")
	m := c.createDept(name, nil)
	if m["name"] != name {
		t.Errorf("want name=%s, got %v", name, m["name"])
	}
	if m["parent_id"] != nil {
		t.Errorf("want parent_id null, got %v", m["parent_id"])
	}
	if m["id"] == nil {
		t.Errorf("want id in response")
	}
}

func TestPostDepartment_ChildCreated(t *testing.T) {
	c := client(t)
	parentID := intID(c.createDept(uniq("Corp"), nil))
	resp := c.do(http.MethodPost, "/departments/", map[string]interface{}{
		"name": uniq("Backend"), "parent_id": parentID,
	})
	m := mustStatus(t, resp, http.StatusCreated)
	if int(m["parent_id"].(float64)) != parentID {
		t.Errorf("want parent_id=%d, got %v", parentID, m["parent_id"])
	}
}

func TestPostDepartment_NameRequired(t *testing.T) {
	resp := client(t).do(http.MethodPost, "/departments/", map[string]interface{}{})
	mustStatus(t, resp, http.StatusBadRequest)
}

func TestPostDepartment_NameWhitespaceRejected(t *testing.T) {
	resp := client(t).do(http.MethodPost, "/departments/", map[string]interface{}{"name": "   "})
	mustStatus(t, resp, http.StatusBadRequest)
}

func TestPostDepartment_NameTrimmed(t *testing.T) {
	c := client(t)
	resp := c.do(http.MethodPost, "/departments/", map[string]interface{}{"name": "  Finance  "})
	m := mustStatus(t, resp, http.StatusCreated)
	c.cleanup(intID(m))
	if m["name"] != "Finance" {
		t.Errorf("want trimmed name 'Finance', got %v", m["name"])
	}
}

func TestPostDepartment_DuplicateNameSameParent(t *testing.T) {
	c := client(t)
	name := uniq("Backend")
	c.createDept(name, nil)
	resp := c.do(http.MethodPost, "/departments/", map[string]interface{}{"name": name})
	mustStatus(t, resp, http.StatusConflict)
}

func TestPostDepartment_SameNameDifferentParentAllowed(t *testing.T) {
	c := client(t)
	idA := intID(c.createDept(uniq("DivA"), nil))
	idB := intID(c.createDept(uniq("DivB"), nil))
	name := uniq("Backend")
	resp1 := c.do(http.MethodPost, "/departments/", map[string]interface{}{"name": name, "parent_id": idA})
	resp2 := c.do(http.MethodPost, "/departments/", map[string]interface{}{"name": name, "parent_id": idB})
	mustStatus(t, resp1, http.StatusCreated)
	mustStatus(t, resp2, http.StatusCreated)
}

func TestPostDepartment_ParentNotFound(t *testing.T) {
	resp := client(t).do(http.MethodPost, "/departments/", map[string]interface{}{
		"name": uniq("Ghost"), "parent_id": 999999,
	})
	mustStatus(t, resp, http.StatusNotFound)
}

// ============================================================
// POST /departments/{id}/employees/
// ============================================================

func TestPostEmployee_Created(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	resp := c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", deptID), map[string]interface{}{
		"full_name": "Alice Johnson",
		"position":  "Engineer",
	})
	m := mustStatus(t, resp, http.StatusCreated)
	if m["full_name"] != "Alice Johnson" {
		t.Errorf("full_name mismatch: %v", m["full_name"])
	}
	if m["hired_at"] != nil {
		t.Errorf("want hired_at null, got %v", m["hired_at"])
	}
}

func TestPostEmployee_WithHiredAt(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	resp := c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", deptID), map[string]interface{}{
		"full_name": "Bob Smith",
		"position":  "Developer",
		"hired_at":  "2023-06-01",
	})
	m := mustStatus(t, resp, http.StatusCreated)
	if m["hired_at"] == nil {
		t.Errorf("want hired_at populated")
	}
}

func TestPostEmployee_InvalidHiredAtFormat(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	resp := c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", deptID), map[string]interface{}{
		"full_name": "Bob", "position": "Dev", "hired_at": "01/06/2023",
	})
	mustStatus(t, resp, http.StatusBadRequest)
}

func TestPostEmployee_DeptNotFound(t *testing.T) {
	resp := client(t).do(http.MethodPost, "/departments/999999/employees/", map[string]interface{}{
		"full_name": "Alice", "position": "Engineer",
	})
	mustStatus(t, resp, http.StatusNotFound)
}

func TestPostEmployee_FullNameRequired(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	resp := c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", deptID), map[string]interface{}{
		"position": "Engineer",
	})
	mustStatus(t, resp, http.StatusBadRequest)
}

func TestPostEmployee_FullNameEmpty(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	resp := c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", deptID), map[string]interface{}{
		"full_name": "   ", "position": "Engineer",
	})
	mustStatus(t, resp, http.StatusBadRequest)
}

func TestPostEmployee_PositionRequired(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	resp := c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", deptID), map[string]interface{}{
		"full_name": "Alice",
	})
	mustStatus(t, resp, http.StatusBadRequest)
}

func TestPostEmployee_PositionEmpty(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	resp := c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", deptID), map[string]interface{}{
		"full_name": "Alice", "position": "   ",
	})
	mustStatus(t, resp, http.StatusBadRequest)
}

// ============================================================
// GET /departments/{id}
// ============================================================

func TestGetDepartment_NotFound(t *testing.T) {
	resp := client(t).do(http.MethodGet, "/departments/999999", nil)
	mustStatus(t, resp, http.StatusNotFound)
}

func TestGetDepartment_ReturnsFields(t *testing.T) {
	c := client(t)
	name := uniq("Engineering")
	deptID := intID(c.createDept(name, nil))
	m := mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d", deptID), nil), http.StatusOK)
	if m["name"] != name {
		t.Errorf("want name=%s, got %v", name, m["name"])
	}
	if int(m["id"].(float64)) != deptID {
		t.Errorf("want id=%d, got %v", deptID, m["id"])
	}
}

func TestGetDepartment_Depth1ExcludesGrandchildren(t *testing.T) {
	c := client(t)
	rootID := intID(c.createDept(uniq("Root"), nil))
	childID := intID(c.createDept(uniq("Child"), ptr(rootID)))
	c.createDept(uniq("Grandchild"), ptr(childID))

	m := mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d?depth=1", rootID), nil), http.StatusOK)
	children := m["children"].([]interface{})
	if len(children) != 1 {
		t.Fatalf("want 1 child, got %d", len(children))
	}
	child := children[0].(map[string]interface{})
	if gc, ok := child["children"].([]interface{}); ok && len(gc) > 0 {
		t.Errorf("depth=1 must not expose grandchildren")
	}
}

func TestGetDepartment_Depth2IncludesGrandchildren(t *testing.T) {
	c := client(t)
	rootID := intID(c.createDept(uniq("Root"), nil))
	childID := intID(c.createDept(uniq("Child"), ptr(rootID)))
	c.createDept(uniq("Grandchild"), ptr(childID))

	m := mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d?depth=2", rootID), nil), http.StatusOK)
	child := m["children"].([]interface{})[0].(map[string]interface{})
	gc, ok := child["children"].([]interface{})
	if !ok || len(gc) != 1 {
		t.Errorf("depth=2 must include grandchildren, got %v", child["children"])
	}
}

func TestGetDepartment_Depth5ReturnsFullChain(t *testing.T) {
	c := client(t)
	cur := intID(c.createDept(uniq("L0"), nil))
	rootID := cur
	for i := 1; i <= 5; i++ {
		cur = intID(c.createDept(uniq(fmt.Sprintf("L%d", i)), ptr(cur)))
	}
	leafID := cur

	m := mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d?depth=5", rootID), nil), http.StatusOK)
	node := m
	for level := 1; level <= 5; level++ {
		kids, ok := node["children"].([]interface{})
		if !ok || len(kids) == 0 {
			t.Fatalf("chain broken at level %d", level)
		}
		node = kids[0].(map[string]interface{})
	}
	if int(node["id"].(float64)) != leafID {
		t.Errorf("want leaf id=%d, got %v", leafID, node["id"])
	}
}

func TestGetDepartment_DepthCappedAt5(t *testing.T) {
	c := client(t)
	// depth=6 should be treated as 5 (max)
	rootID := intID(c.createDept(uniq("Root"), nil))
	resp := c.do(http.MethodGet, fmt.Sprintf("/departments/%d?depth=6", rootID), nil)
	// should not error — server clamps to 5
	mustStatus(t, resp, http.StatusOK)
}

func TestGetDepartment_IncludeEmployeesTrue(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", deptID), map[string]interface{}{
		"full_name": "Alice", "position": "Engineer",
	}).Body.Close()

	m := mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d?include_employees=true", deptID), nil), http.StatusOK)
	emps, ok := m["employees"].([]interface{})
	if !ok || len(emps) != 1 {
		t.Errorf("want 1 employee, got %v", m["employees"])
	}
}

func TestGetDepartment_IncludeEmployeesFalse(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", deptID), map[string]interface{}{
		"full_name": "Alice", "position": "Engineer",
	}).Body.Close()

	m := mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d?include_employees=false", deptID), nil), http.StatusOK)
	if emps, ok := m["employees"].([]interface{}); ok && len(emps) > 0 {
		t.Errorf("want no employees, got %v", emps)
	}
}

func TestGetDepartment_DefaultIncludesEmployees(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", deptID), map[string]interface{}{
		"full_name": "Alice", "position": "Engineer",
	}).Body.Close()

	// no include_employees param → default true
	m := mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d", deptID), nil), http.StatusOK)
	emps, ok := m["employees"].([]interface{})
	if !ok || len(emps) == 0 {
		t.Errorf("default should include employees, got %v", m["employees"])
	}
}

// ============================================================
// PATCH /departments/{id}
// ============================================================

func TestPatchDepartment_Rename(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("OldName"), nil))
	newName := uniq("NewName")
	m := mustStatus(t, c.do(http.MethodPatch, fmt.Sprintf("/departments/%d", deptID), map[string]interface{}{
		"name": newName,
	}), http.StatusOK)
	if m["name"] != newName {
		t.Errorf("want name=%s, got %v", newName, m["name"])
	}
}

func TestPatchDepartment_Move(t *testing.T) {
	c := client(t)
	idA := intID(c.createDept(uniq("ParentA"), nil))
	idB := intID(c.createDept(uniq("ParentB"), nil))
	childID := intID(c.createDept(uniq("Child"), ptr(idA)))

	m := mustStatus(t, c.do(http.MethodPatch, fmt.Sprintf("/departments/%d", childID), map[string]interface{}{
		"parent_id": idB,
	}), http.StatusOK)
	if int(m["parent_id"].(float64)) != idB {
		t.Errorf("want parent_id=%d, got %v", idB, m["parent_id"])
	}
}

func TestPatchDepartment_MoveToRoot(t *testing.T) {
	c := client(t)
	parentID := intID(c.createDept(uniq("Parent"), nil))
	childID := intID(c.createDept(uniq("Child"), ptr(parentID)))

	m := mustStatus(t, c.do(http.MethodPatch, fmt.Sprintf("/departments/%d", childID), map[string]interface{}{
		"parent_id": nil,
	}), http.StatusOK)
	if m["parent_id"] != nil {
		t.Errorf("want parent_id null after move to root, got %v", m["parent_id"])
	}
}

func TestPatchDepartment_NotFound(t *testing.T) {
	resp := client(t).do(http.MethodPatch, "/departments/999999", map[string]interface{}{"name": "X"})
	mustStatus(t, resp, http.StatusNotFound)
}

func TestPatchDepartment_SelfParent(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	resp := c.do(http.MethodPatch, fmt.Sprintf("/departments/%d", deptID), map[string]interface{}{
		"parent_id": deptID,
	})
	mustStatus(t, resp, http.StatusBadRequest)
}

func TestPatchDepartment_CycleRejected(t *testing.T) {
	c := client(t)
	rootID := intID(c.createDept(uniq("Root"), nil))
	childID := intID(c.createDept(uniq("Child"), ptr(rootID)))
	grandID := intID(c.createDept(uniq("Grand"), ptr(childID)))

	// root → grandchild = cycle
	resp := c.do(http.MethodPatch, fmt.Sprintf("/departments/%d", rootID), map[string]interface{}{
		"parent_id": grandID,
	})
	mustStatus(t, resp, http.StatusConflict)
}

func TestPatchDepartment_NameConflictInTargetParent(t *testing.T) {
	c := client(t)
	idA := intID(c.createDept(uniq("ParentA"), nil))
	idB := intID(c.createDept(uniq("ParentB"), nil))
	name := uniq("Backend")
	c.createDept(name, ptr(idB))           // Backend already exists under B
	childID := intID(c.createDept(name, ptr(idA))) // Backend under A

	// move A's Backend under B where Backend already exists
	resp := c.do(http.MethodPatch, fmt.Sprintf("/departments/%d", childID), map[string]interface{}{
		"parent_id": idB,
	})
	mustStatus(t, resp, http.StatusConflict)
}

func TestPatchDepartment_RenameConflictSameParent(t *testing.T) {
	c := client(t)
	parentID := intID(c.createDept(uniq("Parent"), nil))
	name := uniq("Existing")
	c.createDept(name, ptr(parentID))
	otherID := intID(c.createDept(uniq("Other"), ptr(parentID)))

	resp := c.do(http.MethodPatch, fmt.Sprintf("/departments/%d", otherID), map[string]interface{}{
		"name": name,
	})
	mustStatus(t, resp, http.StatusConflict)
}

// ============================================================
// DELETE /departments/{id}
// ============================================================

func TestDeleteDepartment_NotFound(t *testing.T) {
	resp := client(t).do(http.MethodDelete, "/departments/999999?mode=cascade", nil)
	mustStatus(t, resp, http.StatusNotFound)
}

func TestDeleteDepartment_InvalidMode(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	resp := c.do(http.MethodDelete, fmt.Sprintf("/departments/%d?mode=wrong", deptID), nil)
	mustStatus(t, resp, http.StatusBadRequest)
}

func TestDeleteDepartment_CascadeReturns204(t *testing.T) {
	c := client(t)
	// don't register cleanup — we're deleting manually
	resp1 := c.do(http.MethodPost, "/departments/", map[string]interface{}{"name": uniq("Eng")})
	deptID := intID(mustStatus(t, resp1, http.StatusCreated))

	resp := c.do(http.MethodDelete, fmt.Sprintf("/departments/%d?mode=cascade", deptID), nil)
	mustStatus(t, resp, http.StatusNoContent)
}

func TestDeleteDepartment_CascadeDeletesChildrenAndEmployees(t *testing.T) {
	c := client(t)
	resp1 := c.do(http.MethodPost, "/departments/", map[string]interface{}{"name": uniq("Parent")})
	parentID := intID(mustStatus(t, resp1, http.StatusCreated))

	childID := intID(c.createDept(uniq("Child"), ptr(parentID)))
	c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", childID), map[string]interface{}{
		"full_name": "Alice", "position": "Engineer",
	}).Body.Close()

	mustStatus(t, c.do(http.MethodDelete, fmt.Sprintf("/departments/%d?mode=cascade", parentID), nil), http.StatusNoContent)

	mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d", parentID), nil), http.StatusNotFound)
	mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d", childID), nil), http.StatusNotFound)
}

func TestDeleteDepartment_ReassignMovesEmployees(t *testing.T) {
	c := client(t)
	targetID := intID(c.createDept(uniq("Target"), nil))

	resp1 := c.do(http.MethodPost, "/departments/", map[string]interface{}{"name": uniq("ToDelete")})
	delID := intID(mustStatus(t, resp1, http.StatusCreated))

	c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", delID), map[string]interface{}{
		"full_name": "Bob", "position": "Dev",
	}).Body.Close()
	c.do(http.MethodPost, fmt.Sprintf("/departments/%d/employees/", delID), map[string]interface{}{
		"full_name": "Carol", "position": "Lead",
	}).Body.Close()

	mustStatus(t, c.do(http.MethodDelete,
		fmt.Sprintf("/departments/%d?mode=reassign&reassign_to_department_id=%d", delID, targetID),
		nil,
	), http.StatusNoContent)

	mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d", delID), nil), http.StatusNotFound)

	m := mustStatus(t, c.do(http.MethodGet, fmt.Sprintf("/departments/%d?include_employees=true", targetID), nil), http.StatusOK)
	emps, ok := m["employees"].([]interface{})
	if !ok || len(emps) != 2 {
		t.Errorf("want 2 reassigned employees, got %v", m["employees"])
	}
}

func TestDeleteDepartment_ReassignMissingTargetParam(t *testing.T) {
	c := client(t)
	deptID := intID(c.createDept(uniq("Eng"), nil))
	resp := c.do(http.MethodDelete, fmt.Sprintf("/departments/%d?mode=reassign", deptID), nil)
	mustStatus(t, resp, http.StatusBadRequest)
}

func TestDeleteDepartment_ReassignTargetNotFound(t *testing.T) {
	c := client(t)
	resp1 := c.do(http.MethodPost, "/departments/", map[string]interface{}{"name": uniq("Eng")})
	deptID := intID(mustStatus(t, resp1, http.StatusCreated))
	c.cleanup(deptID)

	resp := c.do(http.MethodDelete, fmt.Sprintf("/departments/%d?mode=reassign&reassign_to_department_id=999999", deptID), nil)
	mustStatus(t, resp, http.StatusNotFound)
}
