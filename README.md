# like-service

Post likes for Snaply. Go 1.22, port 8084, own Postgres DB `likes`. Trusts the `X-User-ID` header set by `api-gateway`.

## Environment Variables

| Variable       | Default                                                                   | Description                    |
|----------------|-------------------------------------------------------------------------------|---------------------------------|
| `DATABASE_URL` | `postgres://snaply:snaply_secret@localhost:5432/likes?sslmode=disable`    | PostgreSQL connection string    |
| `KAFKA_BROKERS`| `localhost:29092`                                                         | Comma-separated Kafka broker list |
| `POST_SERVICE_URL` | `http://localhost:8082`                                               | Used to resolve a post's author before publishing `post.liked` |
| `SERVER_PORT`  | `8084`                                                                    | HTTP listen port                |

## Endpoints

| Method | Path                        | Description                                                    |
|--------|------------------------------|-----------------------------------------------------------------|
| POST   | /api/v1/likes/{post_id}      | Caller likes `post_id`                                          |
| DELETE | /api/v1/likes/{post_id}      | Caller unlikes `post_id`                                        |
| GET    | /api/v1/likes/{post_id}/likers | Cursor-paginated list of user IDs who liked `post_id`          |
| POST   | /api/v1/likes/batch           | `{post_ids: [uuid,...]}` → `[{post_id, count, liked}, ...]`, max 100 ids |
| GET    | /health                       | Health check                                                     |

`POST /batch` is the main read path — feeds render many posts at once, so the frontend fetches counts + "did I like this" status for a whole page of posts in a single call instead of two requests per post.

Publishes `post.liked` to Kafka on a successful like (consumed by `notification-service`).

## Running Locally

```bash
cd ../infra && make up
make migrate-up
make run
```