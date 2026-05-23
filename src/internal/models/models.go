package models

import (
    "time"
)


type Department struct {
    ID        int          `json:"id" gorm:"primaryKey;autoIncrement"`
    Name      string       `json:"name" gorm:"column:name;type:varchar(200);not null;check:name != ''"`
    ParentID  *int         `json:"parent_id,omitempty" gorm:"column:parent_id;index:idx_departments_parent_id"`
    CreatedAt time.Time    `json:"created_at" gorm:"column:created_at;autoCreateTime"`
    Children  []Department `json:"children,omitempty" gorm:"-"`           // не хранится в БД, только для ответов
    Employees []Employee   `json:"employees,omitempty" gorm:"foreignKey:DepartmentID"` // для Preload
}

func (Department) TableName() string {
    return "departments"
}

type Employee struct {
    ID           int        `json:"id" gorm:"primaryKey;autoIncrement"`
    DepartmentID int        `json:"department_id,omitempty" gorm:"column:department_id;not null;index:idx_employee_department"`
    FullName     string     `json:"full_name" gorm:"column:full_name;type:varchar(200);not null;check:full_name != ''"`
    Position     string     `json:"position" gorm:"column:position;type:varchar(200);not null;check:position != ''"`
    HiredAt      *time.Time `json:"hired_at,omitempty" gorm:"column:hired_at;type:date"`
    CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (Employee) TableName() string {
    return "employees"
}
// type Department struct {
// 	ID        int       `gorm:"primaryKey"`
// 	Name      string    `gorm:"type:varchar(200);not null"`
// 	ParentID  *int      `gorm:"index"`
// 	CreatedAt time.Time `gorm:"autoCreateTime"`
// }

// type Employee struct {
// 	ID           int        `gorm:"primaryKey"`
// 	DepartmentID int        `gorm:"not null;index"`
// 	FullName     string     `gorm:"type:varchar(200);not null"`
// 	Position     string     `gorm:"type:varchar(200);not null"`
// 	HiredAt      *time.Time `gorm:"type:date"`      // внимание. тут должна быть просто дата в формате YY-MM-DD
// 	CreatedAt    time.Time  `gorm:"autoCreateTime"` // а тут и везде в остальных time должен быть полный timestamp
// }


