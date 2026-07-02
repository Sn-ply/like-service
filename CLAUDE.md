# like-service

Post likes for Snaply. Go 1.22, port 8084, own Postgres DB `likes`. Trusts the `X-User-ID` header set by `api-gateway`.

## Layout

`cmd/main.go` → `internal/handler` (`like.go`, `routes.go`) → `internal/service` (`like_service.go`) → `internal/repository` (`postgres/like.go`) → `internal/model`

## Endpoints

| Method | Path                         | Description                                                    |
|--------|-------------------------------|-------------------------------------------------------------------|
| POST   | /api/v1/likes/{post_id}       | Caller likes `post_id`                                            |
| DELETE | /api/v1/likes/{post_id}       | Caller unlikes `post_id`                                          |
| GET    | /api/v1/likes/{post_id}/likers | Cursor-paginated list of user IDs who liked `post_id`             |
| POST   | /api/v1/likes/batch            | `{post_ids: [uuid,...]}` → `[{post_id, count, liked}, ...]`, max 100 |
| GET    | /health                        | Health check                                                       |

## Conventions

- Mirrors `relation-service`: IDs only (no post/user identity data stored), `Like()` idempotent (`ON CONFLICT DO NOTHING`), `Unlike()` a no-op if not liked, same cursor pagination scheme.
- Publishes `post.liked` to Kafka on a successful like (consumed by `notification-service`), async in a goroutine. Since this service stores only `(post_id, user_id)` and has no idea who authored the post, publishing first requires an HTTP call to `post-service`'s `GET /posts/{id}` to resolve `author_id` (`internal/client/postclient.go` — TODO there to replace with gRPC). Publish and lookup failures are logged, never surfaced to the caller. No event is published for unlike.
- `/batch` is the primary read path — a feed page renders many posts at once, so the frontend fetches counts + "did I like this" for the whole page in one call instead of two requests per post. `CountsAndStatus` runs two bounded queries (a `GROUP BY` for counts, a small `IN` lookup for the caller's own likes) rather than pulling every like row for popular posts.
- `likes` table: composite PK `(post_id, user_id)`, no self-like restriction needed (unlike follows) since liking is symmetric with the post, not another user.

## Running

```bash
cd ../infra && make up
make migrate-up
make run
```

Default `DATABASE_URL`: `postgres://snaply:snaply_secret@localhost:5432/likes?sslmode=disable`
