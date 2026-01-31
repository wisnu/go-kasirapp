-- DDL
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP SEQUENCE IF EXISTS categories_id_seq;
DROP SEQUENCE IF EXISTS products_id_seq;

CREATE SEQUENCE categories_id_seq START 1;
CREATE SEQUENCE products_id_seq START 1;

CREATE TABLE categories (
    id INTEGER NOT NULL DEFAULT nextval('categories_id_seq') PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE products (
    id INTEGER NOT NULL DEFAULT nextval('products_id_seq') PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    price NUMERIC(12, 2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL CHECK (stock >= 0),
    category_id INTEGER REFERENCES categories(id)
);

-- DML (seed data)
INSERT INTO categories (id, name, description) VALUES
    (1, 'Electronics', 'Electronic devices and gadgets'),
    (2, 'Accessories', 'Related accessories and add-ons');

INSERT INTO products (id, name, price, stock, category_id) VALUES
    (1, 'Laptop', 999.99, 10, 1),
    (2, 'Smartphone', 499.99, 25, 1),
    (3, 'Tablet', 299.99, 15, 1),
    (4, 'Headphones', 99.99, 60, 2);

-- Sync sequence ke angka tertinggi setelah seed
SELECT setval('categories_id_seq', (SELECT MAX(id) FROM categories));
SELECT setval('products_id_seq', (SELECT MAX(id) FROM products));
