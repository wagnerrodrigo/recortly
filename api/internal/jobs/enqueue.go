package jobs

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const (
	QueueRenders     = "renders"
	QueuePublish     = "publish"
	QueueMaintenance = "maintenance"

	TypeRender  = "render:run"
	TypePublish = "publish:run"
	TypeCleanup = "maintenance:cleanup"
)

type RenderPayload struct {
	RenderID int64 `json:"renderId"`
}

type PublishPayload struct {
	PostID int64 `json:"postId"`
}

type Enqueuer struct {
	Client *asynq.Client
}

func NewEnqueuer(redisAddr string) *Enqueuer {
	return &Enqueuer{Client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})}
}

func (e *Enqueuer) Close() error { return e.Client.Close() }

func (e *Enqueuer) EnqueueRender(renderID int64) error {
	raw, _ := json.Marshal(RenderPayload{RenderID: renderID})
	_, err := e.Client.Enqueue(asynq.NewTask(TypeRender, raw),
		asynq.Queue(QueueRenders),
		asynq.MaxRetry(1),
		asynq.Timeout(3*time.Hour))
	return err
}

func (e *Enqueuer) EnqueuePublish(postID int64, at time.Time) error {
	raw, _ := json.Marshal(PublishPayload{PostID: postID})
	opts := []asynq.Option{asynq.Queue(QueuePublish), asynq.MaxRetry(2), asynq.Timeout(30 * time.Minute)}
	if at.After(time.Now()) {
		opts = append(opts, asynq.ProcessAt(at))
	}
	_, err := e.Client.Enqueue(asynq.NewTask(TypePublish, raw), opts...)
	return err
}
