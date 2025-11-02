# API Endpoint Documentation

## GET /api/accounts

Retrieves accounts from the database. Supports optional deletion when count parameter is provided.

### Authentication
Requires Bearer token in Authorization header.

### Parameters
- `count` (optional, integer): Number of accounts to retrieve and delete
  - If `count > 0`: Returns specified number of accounts and deletes them from database
  - If `count = 0` or not provided: Returns all accounts without deletion

### Request Examples

**Get and delete 1 account:**
```bash
curl -H "Authorization: Bearer your-token" \
  "http://localhost:8080/api/accounts?count=1"
```

**Get and delete 10 accounts:**
```bash
curl -H "Authorization: Bearer your-token" \
  "http://localhost:8080/api/accounts?count=10"
```

**View all accounts (no deletion):**
```bash
curl -H "Authorization: Bearer your-token" \
  "http://localhost:8080/api/accounts"
```

### Response
Returns JSON array of account objects:
```json
[
  {
    "id": 1,
    "refresh_token": "...",
    "client_id": "...",
    "client_secret": "..."
  }
]
```

### Behavior
When `count` parameter is provided:
1. Fetches specified number of accounts
2. Deletes them from database
3. Updates metrics (current_count, used_count, api_call_count)
4. Returns the deleted accounts

### Error Responses
- `401 Unauthorized`: Missing or invalid Bearer token
- `400 Bad Request`: Invalid count parameter (non-numeric or negative)
  ```json
  {
    "error": "invalid count parameter"
  }
  ```
- `400 Bad Request`: Requested count exceeds available accounts
  ```json
  {
    "error": "insufficient accounts",
    "available": 5,
    "requested": 10
  }
  ```
- `200 OK`: Success (returns empty array if no accounts available)
