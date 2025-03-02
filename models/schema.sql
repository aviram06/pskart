CREATE TABLE orders (
    order_id VARCHAR(255) PRIMARY KEY,
	user_id VARCHAR(255),
	item_ids INTEGER[],
	total_amount DECIMAL
);