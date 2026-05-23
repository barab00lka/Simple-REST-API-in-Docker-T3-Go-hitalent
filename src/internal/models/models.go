package models

import (
    "time"
)


type Department struct {
    ID        int          `json:"id" gorm:"primaryKey;autoIncrement"`
    Name      string       `json:"name" gorm:"column:name;type:varchar(200);not null;check:name != ''"`
    ParentID  *int         `json:"parent_id,omitempty" gorm:"column:parent_id;index:idx_departments_parent_id"`
    CreatedAt time.Time    `json:"created_at" gorm:"column:created_at;autoCreateTime"`
    Children  []Department `json:"children,omitempty" gorm:"-"`
    Employees []Employee   `json:"employees,omitempty" gorm:"foreignKey:DepartmentID"`
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
