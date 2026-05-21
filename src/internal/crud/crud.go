package crud

import (
	"context"
	"errors"
	"time"

	dbmodel "main/internal/models"
	"gorm.io/gorm"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrCycle      = errors.New("cycle in department tree")
	ErrSelfParent = errors.New("department cannot be its own parent")
)

// ---------- Generic helpers ----------

func gFirst[T any](ctx context.Context, db *gorm.DB, cond string, arg any) (T, error) {
	v, err := gorm.G[T](db).Where(cond, arg).First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		var zero T
		return zero, ErrNotFound
	}
	return v, err
}

func gExists[T any](ctx context.Context, db *gorm.DB, id int) error {
	_, err := gFirst[T](ctx, db, "id = ?", id)
	return err
}

// ---------- Department ----------

func CreateDepartment(ctx context.Context, db *gorm.DB, name string, parentID *int) (*dbmodel.Department, error) {
	if parentID != nil {
		if err := gExists[dbmodel.Department](ctx, db, *parentID); err != nil {
			return nil, err
		}
	}
	if err := checkNameUnique(db, name, parentID, 0); err != nil {
		return nil, err
	}
	dept := dbmodel.Department{Name: name, ParentID: parentID}
	if err := gorm.G[dbmodel.Department](db).Create(ctx, &dept); err != nil {
		return nil, err
	}
	return &dept, nil
}

func GetDepartment(ctx context.Context, db *gorm.DB, id, depth int, includeEmployees bool) (*dbmodel.Department, error) {
	dept, err := gFirst[dbmodel.Department](ctx, db, "id = ?", id)
	if err != nil {
		return nil, err
	}
	if includeEmployees {
		emps, err := gorm.G[dbmodel.Employee](db).
			Raw("SELECT * FROM employees WHERE department_id = ? ORDER BY created_at", id).
			Find(ctx)
		if err != nil {
			return nil, err
		}
		dept.Employees = emps
	}
	if depth > 0 {
		if err := loadChildren(ctx, db, &dept, depth-1, includeEmployees); err != nil {
			return nil, err
		}
	}
	return &dept, nil
}

func loadChildren(ctx context.Context, db *gorm.DB, dept *dbmodel.Department, depth int, includeEmployees bool) error {
	children, err := gorm.G[dbmodel.Department](db).Where("parent_id = ?", dept.ID).Find(ctx)
	if err != nil {
		return err
	}
	for i := range children {
		if includeEmployees {
			emps, err := gorm.G[dbmodel.Employee](db).
				Raw("SELECT * FROM employees WHERE department_id = ? ORDER BY created_at", children[i].ID).
				Find(ctx)
			if err != nil {
				return err
			}
			children[i].Employees = emps
		}
		if depth > 0 {
			if err := loadChildren(ctx, db, &children[i], depth-1, includeEmployees); err != nil {
				return err
			}
		}
	}
	dept.Children = children
	return nil
}

func PatchDepartment(ctx context.Context, db *gorm.DB, id int, name *string, setParent bool, parentID *int) (*dbmodel.Department, error) {
	dept, err := gFirst[dbmodel.Department](ctx, db, "id = ?", id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		dept.Name = *name
	}
	if setParent {
		if parentID != nil {
			if *parentID == id {
				return nil, ErrSelfParent
			}
			if err := gExists[dbmodel.Department](ctx, db, *parentID); err != nil {
				return nil, err
			}
			if isCyclic(ctx, db, id, *parentID) {
				return nil, ErrCycle
			}
		}
		dept.ParentID = parentID
	}

	if err := checkNameUnique(db, dept.Name, dept.ParentID, id); err != nil {
		return nil, err
	}
	
	// Updates со struct пропускает nil-поля — передаём dept для name/parent_id != nil.
	if _, err := gorm.G[dbmodel.Department](db).Where("id = ?", id).Updates(ctx, dept); err != nil {
		return nil, err
	}
	// NULL parent_id GORM пропустит в struct update — обновляем отдельно.
	if setParent && parentID == nil {
		if _, err := gorm.G[dbmodel.Department](db).
			Where("id = ?", id).
			Update(ctx, "parent_id", gorm.Expr("NULL")); err != nil {
			return nil, err
		}
	}

	return &dept, nil
}

func DeleteDepartment(ctx context.Context, db *gorm.DB, id int, mode string, reassignTo *int) error {
	if err := gExists[dbmodel.Department](ctx, db, id); err != nil {
		return err
	}
	if mode == "reassign" {
		if err := gExists[dbmodel.Department](ctx, db, *reassignTo); err != nil {
			return err
		}
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if mode == "cascade" {
			return deleteCascade(ctx, tx, id)
		}
		return deleteReassign(ctx, tx, id, *reassignTo)
	})
}

func deleteCascade(ctx context.Context, db *gorm.DB, id int) error {
	children, err := gorm.G[dbmodel.Department](db).Where("parent_id = ?", id).Find(ctx)
	if err != nil {
		return err
	}
	for _, child := range children {
		if err := deleteCascade(ctx, db, child.ID); err != nil {
			return err
		}
	}
	if _, err := gorm.G[dbmodel.Employee](db).Where("department_id = ?", id).Delete(ctx); err != nil {
		return err
	}
	_, err = gorm.G[dbmodel.Department](db).Where("id = ?", id).Delete(ctx) 

	return  err
}

func deleteReassign(ctx context.Context, db *gorm.DB, id, reassignTo int) error {
	if _, err := gorm.G[dbmodel.Employee](db).
		Where("department_id = ?", id).
		Update(ctx, "department_id", reassignTo); err != nil {
		return err
	}

	dept, err := gFirst[dbmodel.Department](ctx, db, "id = ?", id)
	if err != nil {
		return err
	}

	var newParent any
	if dept.ParentID == nil {
		newParent = gorm.Expr("NULL")
	} else {
		newParent = *dept.ParentID
	}
	if _, err := gorm.G[dbmodel.Department](db).
		Where("parent_id = ?", id).
		Update(ctx, "parent_id", newParent); err != nil {
		return err
	}

	_,  err = gorm.G[dbmodel.Department](db).Where("id = ?", id).Delete(ctx)
	return err
}

// ---------- Employee ----------

func AddEmployee(ctx context.Context, db *gorm.DB, deptID int, fullName, position string, hiredAt *time.Time) (*dbmodel.Employee, error) {
	if err := gExists[dbmodel.Department](ctx, db, deptID); err != nil {
		return nil, err
	}
	emp := dbmodel.Employee{
		DepartmentID: deptID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      hiredAt,
	}
	if err := gorm.G[dbmodel.Employee](db).Create(ctx, &emp); err != nil {
		return nil, err
	}
	return &emp, nil
}

// ---------- Internal helpers ----------

// checkNameUnique использует старый API — Count в generic API не задокументирован.
func checkNameUnique(db *gorm.DB, name string, parentID *int, excludeID int) error {
	q := db.Model(new(dbmodel.Department)).Where("name = ?", name)
	if parentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *parentID)
	}
	if excludeID != 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrConflict
	}
	return nil
}

// isCyclic возвращает true если candidateParent является потомком id.
func isCyclic(ctx context.Context, db *gorm.DB, id, candidateParent int) bool {
	seen := map[int]bool{}
	cur := candidateParent
	for {
		if cur == id {
			return true
		}
		if seen[cur] {
			return false
		}
		seen[cur] = true
		d, err := gFirst[dbmodel.Department](ctx, db, "id = ?", cur)
		if err != nil {
			return false
		}
		if d.ParentID == nil {
			return false
		}
		cur = *d.ParentID
	}
}
