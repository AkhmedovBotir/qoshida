package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ActionCreate  = "create"
	ActionUpdate  = "update"
	ActionApprove = "approve"
	ActionReject  = "reject"

	TypeAdmin        = "admin"
	TypeManager      = "manager"
	TypeShopDirector = "shop_director"
	TypeLocalShop    = "local_shop"
	TypeSeller       = "seller"
	TypeKontragent       = "kontragent"
	TypeServiceProvider  = "service_provider"
)

type contextKey string

const actorKey contextKey = "audit_actor"

type Actor struct {
	Type string
	ID   uuid.UUID
	Name string
	Role string
}

type Entry struct {
	Name string `json:"name"`
	Role string `json:"role"`
	At   string `json:"at"`
}

type Trail struct {
	Created  *Entry  `json:"created,omitempty"`
	Approved *Entry  `json:"approved,omitempty"`
	Rejected *Entry  `json:"rejected,omitempty"`
	Updates  []Entry `json:"updates,omitempty"`
}

type row struct {
	EntityID  uuid.UUID
	Action    string
	ActorType string
	ActorID   uuid.UUID
	ActorName string
	ActorRole string
	At        time.Time
}

var defaultPool *pgxpool.Pool

func Init(pool *pgxpool.Pool) {
	defaultPool = pool
}

func With(ctx context.Context, actor Actor) context.Context {
	if actor.ID == uuid.Nil || actor.Name == "" {
		return ctx
	}
	return context.WithValue(ctx, actorKey, actor)
}

func From(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(actorKey).(Actor)
	return actor, ok && actor.ID != uuid.Nil
}

func AfterCreate(ctx context.Context, entity string, id uuid.UUID) *Trail {
	Stamp(ctx, entity, id, ActionCreate)
	return Bind(ctx, entity, id)
}

func AfterUpdate(ctx context.Context, entity string, id uuid.UUID) *Trail {
	Stamp(ctx, entity, id, ActionUpdate)
	return Bind(ctx, entity, id)
}

func AfterApprove(ctx context.Context, entity string, id uuid.UUID) *Trail {
	Stamp(ctx, entity, id, ActionApprove)
	return Bind(ctx, entity, id)
}

func AfterReject(ctx context.Context, entity string, id uuid.UUID) *Trail {
	Stamp(ctx, entity, id, ActionReject)
	return Bind(ctx, entity, id)
}

func Stamp(ctx context.Context, entity string, id uuid.UUID, action string) {
	if defaultPool == nil || entity == "" || id == uuid.Nil {
		return
	}
	actor, ok := From(ctx)
	if !ok {
		return
	}
	_, _ = defaultPool.Exec(ctx, `
		INSERT INTO record_audits (entity_type, entity_id, action, actor_type, actor_id, actor_name, actor_role)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entity, id, action, actor.Type, actor.ID, actor.Name, actor.Role)
}

func Bind(ctx context.Context, entity string, id uuid.UUID) *Trail {
	m := BindMany(ctx, entity, []uuid.UUID{id})
	t, ok := m[id]
	if !ok {
		return nil
	}
	return &t
}

func BindMany(ctx context.Context, entity string, ids []uuid.UUID) map[uuid.UUID]Trail {
	out := map[uuid.UUID]Trail{}
	if defaultPool == nil || entity == "" || len(ids) == 0 {
		return out
	}
	rows, err := defaultPool.Query(ctx, `
		SELECT entity_id, action, actor_type, actor_id, actor_name, actor_role, created_at
		FROM record_audits
		WHERE entity_type = $1 AND entity_id = ANY($2)
		ORDER BY created_at ASC
	`, entity, ids)
	if err != nil {
		return out
	}
	defer rows.Close()

	grouped := map[uuid.UUID][]row{}
	for rows.Next() {
		var item row
		if err = rows.Scan(&item.EntityID, &item.Action, &item.ActorType, &item.ActorID, &item.ActorName, &item.ActorRole, &item.At); err != nil {
			continue
		}
		grouped[item.EntityID] = append(grouped[item.EntityID], item)
	}
	for id, list := range grouped {
		out[id] = summarize(list)
	}
	return out
}

func summarize(list []row) Trail {
	var trail Trail
	seenUpdate := map[string]int{}
	var updates []Entry
	for _, item := range list {
		entry := Entry{Name: item.ActorName, Role: item.ActorRole, At: item.At.UTC().Format(time.RFC3339)}
		if item.Action == ActionCreate && trail.Created == nil {
			trail.Created = &entry
			continue
		}
		if item.Action == ActionApprove {
			trail.Approved = &entry
			continue
		}
		if item.Action == ActionReject {
			trail.Rejected = &entry
			continue
		}
		if item.Action != ActionUpdate {
			continue
		}
		key := item.ActorType + ":" + item.ActorID.String()
		if idx, ok := seenUpdate[key]; ok {
			updates[idx] = entry
			continue
		}
		seenUpdate[key] = len(updates)
		updates = append(updates, entry)
	}
	if n := len(updates); n > 3 {
		updates = updates[n-3:]
	}
	trail.Updates = updates
	return trail
}

func RoleAdmin() string        { return "Admin" }
func RoleRegionManager() string { return "Viloyat menejeri" }
func RoleDistrictManager() string {
	return "Tuman menejeri"
}
func RoleDirector() string   { return "Savdo uyi rahbari" }
func RoleShop() string       { return "Mahalla do‘koni" }
func RoleSeller() string     { return "Sotuvchi" }
func RoleKontragent() string      { return "Kontragent" }
func RoleServiceProvider() string { return "Xizmat ko‘rsatuvchi" }
