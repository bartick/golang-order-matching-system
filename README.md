# Golang Order Matching System
This repository contains a simple order matching system implemented in Golang. It is designed to handle buy and sell orders, match them based on price and time priority, and execute trades accordingly.

## Startup

To start the system make sure you have go1.24.3 installed and run the following command:

```bash
go mod tidy
go run main.go
```

Using docker you can run the following command:

```bash
docker build -t order-matching-system .
docker run --rm -it --env-file .env -p 8080:8080 order-matching-system
```

Or using docker-compose:

```bash
docker compose up
```

**NOTE**: I am using postgres as the database, so make sure you have it running and the connection string is set in the `.env` file. And load the migration file from `migrations/` directory to create the necessary tables in the database. If you are using docker compose you will not need to do this as the migration will be run automatically when the container starts.

## API Endpoints
### Create Order
- **Endpoint**: `/order`
- **Method**: `POST`
- **Description**: Create a new order (buy or sell).
- **Curl Example**:
  ```bash
  curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "AAPL",
    "side": "sell",
    "type": "limit",
    "price": 190,
    "quantity": 100
  }'
  ```
- **Request Body**:
  ```json
  {
    "symbol": "string",
    "side": "buy" | "sell",
    "type": "limit" | "market",
    "price": 0.0,
    "quantity": 0
  }
  ```
- **Response**:
    ```json
    {
        "order": {
            "id": "string",
            "symbol": "string",
            "side": "sell" | "buy",
            "type": "limit" | "market",
            "price": 0,
            "initial_quantity": 0,
            "remaining_quantity": 0,
            "status": "open" | "filled" | "cancelled" | "partially_filled",
            "created_at": "2025-06-10T18:27:49.303527Z",
            "updated_at": "2025-06-10T18:27:49.303527Z"
        },
        "trades": [
            {
                "id": "string",
                "buy_order_id": "string",
                "sell_order_id": "string",
                "symbol": "string",
                "price": 0.0,
                "quantity": 0.0,
                "executed_at": "2025-06-10T18:27:49.303527Z"
            }
        ]
    }
    ```

###  Get Order
- **Endpoint**: `/order/{id}`
- **Method**: `GET`
- **Description**: Retrieve details of a specific order by ID.
- **Curl Example**:
  ```bash
  curl -X GET http://localhost:8080/order/636eaccf-68e2-4926-98d7-9897a9bc92b3
  ```
- **Response**:
  ```json
  {
    "id": "string",
    "symbol": "string",
    "side": "sell | buy",
    "type": "limit | market",
    "price": 0.0,
    "initial_quantity": 0.0,
    "remaining_quantity": 0.0,
    "status": "open | filled | cancelled | partially_filled",
    "created_at": "2025-06-10T18:27:49.303527Z",
    "updated_at": "2025-06-10T18:27:49.303527Z"
  }
  ```

### Cancel Order
- **Endpoint**: `/order/{id}`
- **Method**: `DELETE`
- **Description**: Cancel an existing order by ID.
- **Curl Example**:
  ```bash
  curl -X DELETE http://localhost:8080/order/636eaccf-68e2-4926-98d7-9897a9bc92b3
  ```
- **Response**:
  ```json
  {
    "message": "Order cancelled successfully",
    "order": {
      "id": "string",
      "symbol": "string",
      "side": "sell | buy",
      "type": "limit | market",
      "price": 0.0,
      "initial_quantity": 0.0,
      "remaining_quantity": 0.0,
      "status": "cancelled",
      "created_at": "2025-06-10T18:27:49.303527Z",
      "updated_at": "2025-06-10T18:27:49.303527Z"
    }
  }
  ```

### Get Order Book
- **Endpoint**: `/orderbook`
- **Method**: `GET`
- **Description**: Retrieve the current order book.
- **Curl Example**:
  ```bash
  curl -X GET http://localhost:8080/orderbook?symbol=GOOGL
  ```
- **Query Parameters**:
    - `symbol`: The stock symbol for which to retrieve the order book (e.g., `AAPL`, `GOOGL`).
- **Response**:
  ```json
  {
    "symbol": "string",
    "bids": [
        {
            "price": 0,
            "total_quantity": 0,
            "order_count": 0
        }
    ],
    "asks": [
        {
            "price": 0,
            "total_quantity": 0,
            "order_count": 0
        }
    ]
  }
  ```
### Get Trades
- **Endpoint**: `/trades`
- **Method**: `GET`
- **Description**: Retrieve the list of executed trades.
- **Curl Example**:
  ```bash
  curl -X GET http://localhost:8080/trades?symbol=AAPL
  ```
- **Query Parameters**:
    - `symbol`: The stock symbol for which to retrieve trades (e.g., `AAPL`, `GOOGL`).
- **Response**:
  ```json
  {
    "trades": [
        {
            "id": "string",
            "buy_order_id": "string",
            "sell_order_id": "string",
            "symbol": "string",
            "price": 0,
            "quantity": 0,
            "executed_at": "2025-06-10T18:33:22.070698Z"
        }
    ]
  }
  ```