-- SCHEMA: public

-- DROP SCHEMA IF EXISTS public ;

CREATE SCHEMA IF NOT EXISTS public
    AUTHORIZATION pg_database_owner;

COMMENT ON SCHEMA public
    IS 'standard public schema';

GRANT USAGE ON SCHEMA public TO PUBLIC;

GRANT ALL ON SCHEMA public TO pg_database_owner;

CREATE TABLE departments (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name varchar(200) NOT NULL CHECK (name != ''),
    parent_id INTEGER NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE departments
ADD CONSTRAINT fk_department_parent
FOREIGN KEY (parent_id) REFERENCES departments(id);

ALTER TABLE departments
ADD CONSTRAINT parent_not_self 
CHECK (parent_id != id);

-- Индекс для ускорения запросов по parent_id (часто нужен для построения дерева)
CREATE INDEX idx_departments_parent_id ON departments(parent_id);
-- Индексы обеспечивающие уникальность имен предприятий:
-- На верхнем уровне
CREATE UNIQUE INDEX idx_unique_dept_name_root ON departments (name) WHERE parent_id IS NULL;
-- У дочерних предприятий в рамках одного родительского
CREATE UNIQUE INDEX idx_unique_dept_name_per_parent ON departments (parent_id, name) WHERE parent_id IS NOT NULL;

CREATE TABLE employees (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    department_id INTEGER NOT NULL,
    full_name varchar(200) NOT NULL CHECK (full_name != ''),
    position varchar(200) NOT NULL CHECK (position != ''),
    hired_at DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE employees
ADD CONSTRAINT employee_department_id_fkey
FOREIGN KEY (department_id) REFERENCES departments(id);
