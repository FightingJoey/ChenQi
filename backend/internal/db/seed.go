package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type seedItem struct {
	StepKey string
	Title   string
	Body    string
}

var seedContent = []seedItem{
	{"silence", "静坐呼吸", "找一个安静的地方坐下，闭上眼睛。用鼻子缓慢吸气 4 秒，屏息 2 秒，再缓缓呼气 6 秒。把注意力放在呼吸上，若思绪飘走，温柔地拉回。"},
	{"silence", "感恩静默", "静静坐着，在心里默念今天想感谢的三件事。不评判、不分析，只感受感恩带来的平静。"},
	{"affirmations", "成长肯定语", "我每天都在变得更好。我有能力创造我想要的生活。我值得成功、健康与幸福。挑战让我更强大。"},
	{"affirmations", "专注与能量", "我醒来就充满能量。我专注于当下最重要的事。我的行动与目标一致。我选择积极与感恩。"},
	{"visualization", "理想早晨", "闭上眼睛，想象理想中的自己已经实现了目标：清晰看见场景、听到声音、感受情绪。想象自己自信地完成今天最重要的三件事。"},
	{"visualization", "目标达成日", "想象一年后的今天：你已经达成了最重要的目标。细节越具体越好——环境、穿着、身边的人、你的感受。在脑中「预演」成功。"},
	{"exercise", "七点拉伸", "1) 颈部左右侧屈各 20 秒\n2) 肩部绕环 20 次\n3) 体前屈 30 秒\n4) 弓步拉伸左右各 30 秒\n5) 平板支撑 30–60 秒\n6) 开合跳或原地慢跑 1 分钟\n7) 深呼吸收尾"},
	{"exercise", "晨间活力操", "做 3 轮：深蹲 15 次 → 俯卧撑（跪姿亦可）10 次 → 平板支撑 30 秒 → 高抬腿 30 秒。量力而行，保持呼吸顺畅。"},
	{"reading", "今日摘录 · 习惯", "「我们重复做的事成就了我们。卓越不是一次行动，而是一种习惯。」——亚里士多德\n\n花几分钟阅读你正在读的书，或反思这句话对你今天的行动意味着什么。"},
	{"reading", "今日摘录 · 早起", "「你如何度过早晨，往往决定了你如何度过一天；你如何度过一天，最终决定了你如何度过一生。」——Hal Elrod\n\n继续阅读你正在读的个人成长类书籍至少 10 页。"},
	{"scribing", "三问日记", "请写下：\n1. 我今天最感恩的三件事是什么？\n2. 今天最重要的一件事是什么？\n3. 我想成为怎样的人，今天可以迈出哪一小步？"},
	{"scribing", "自由书写", "不计对错，连续书写 5–10 分钟。写下此刻的想法、情绪、担忧或灵感。写完后可以圈出一两个行动项。"},
}

func Seed(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM content_items WHERE user_id IS NULL`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	now := time.Now().UTC()
	for _, item := range seedContent {
		_, err := db.Exec(
			`INSERT INTO content_items (id, user_id, step_key, title, body, is_default, created_at) VALUES (?, NULL, ?, ?, ?, 1, ?)`,
			uuid.NewString(), item.StepKey, item.Title, item.Body, now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
