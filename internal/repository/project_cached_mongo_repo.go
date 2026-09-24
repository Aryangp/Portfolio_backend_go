package repository

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"portfolio-backend/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CachedMongoProjectRepository provides a high-performance in-memory cached view over a MongoDB projects collection.
// All read queries are served directly from memory; on a cache miss or expiration, the complete project set is fetched from MongoDB.
type CachedMongoProjectRepository struct {
	collection  *mongo.Collection
	mu          sync.RWMutex
	projects    []model.Project
	byID        map[string]model.Project
	bySlug      map[string]model.Project
	lastFetched time.Time
	ttl         time.Duration
}

// NewCachedMongoProjectRepository creates a new CachedMongoProjectRepository and warms the cache directly from MongoDB.
func NewCachedMongoProjectRepository(collection *mongo.Collection, ttl time.Duration) (*CachedMongoProjectRepository, error) {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	repo := &CachedMongoProjectRepository{
		collection: collection,
		byID:       make(map[string]model.Project),
		bySlug:     make(map[string]model.Project),
		ttl:        ttl,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Initial cache warm-up directly from MongoDB
	if err := repo.refreshCache(ctx); err != nil {
		log.Printf("⚠️ Initial MongoDB project cache warm-up failed: %v", err)
	} else {
		log.Printf("✅ MongoDB project cache warmed with %d projects from database", len(repo.projects))
	}

	return repo, nil
}

// refreshCache fetches all projects from MongoDB and populates the in-memory cache
func (r *CachedMongoProjectRepository) refreshCache(ctx context.Context) error {
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "level", Value: -1}, {Key: "createdAt", Value: -1}}))
	if err != nil {
		return fmt.Errorf("failed to query projects from mongo: %w", err)
	}
	defer cursor.Close(ctx)

	var fetched []model.Project
	if err := cursor.All(ctx, &fetched); err != nil {
		return fmt.Errorf("failed to decode mongo projects: %w", err)
	}

	newByID := make(map[string]model.Project, len(fetched))
	newBySlug := make(map[string]model.Project, len(fetched))

	for _, p := range fetched {
		newByID[p.ID] = p
		if p.Slug != "" {
			newBySlug[strings.ToLower(p.Slug)] = p
		}
	}

	r.mu.Lock()
	r.projects = fetched
	r.byID = newByID
	r.bySlug = newBySlug
	r.lastFetched = time.Now()
	r.mu.Unlock()

	return nil
}

func (r *CachedMongoProjectRepository) isStale() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.projects == nil || time.Since(r.lastFetched) > r.ttl
}

// GetAll returns projects from cache, optionally filtered by category, tag, or featured status
func (r *CachedMongoProjectRepository) GetAll(category, tag string, featured *bool) ([]model.Project, error) {
	if r.isStale() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = r.refreshCache(ctx)
		cancel()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []model.Project
	for _, p := range r.projects {
		if category != "" && !strings.EqualFold(p.Category, category) && !strings.EqualFold(p.BoxCategory, category) {
			continue
		}
		if tag != "" {
			tagMatched := false
			for _, t := range p.TechStack {
				if strings.EqualFold(t, tag) {
					tagMatched = true
					break
				}
			}
			if !tagMatched {
				continue
			}
		}
		if featured != nil && p.Featured != *featured {
			continue
		}
		result = append(result, p)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Level != result[j].Level {
			return result[i].Level > result[j].Level
		}
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	if result == nil {
		result = []model.Project{}
	}
	return result, nil
}

// GetByID looks up a project from cache. On a cache miss, it re-queries MongoDB and refreshes cache.
func (r *CachedMongoProjectRepository) GetByID(id string) (*model.Project, error) {
	if r.isStale() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = r.refreshCache(ctx)
		cancel()
	}

	r.mu.RLock()
	p, found := r.byID[id]
	r.mu.RUnlock()

	if found {
		return &p, nil
	}

	// Cache Miss: Query full dataset from Mongo to refresh cache
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.refreshCache(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	p, found = r.byID[id]
	r.mu.RUnlock()

	if !found {
		return nil, ErrProjectNotFound
	}
	return &p, nil
}

// GetBySlug looks up a project by slug from cache. On a cache miss, it re-queries MongoDB and refreshes cache.
func (r *CachedMongoProjectRepository) GetBySlug(slug string) (*model.Project, error) {
	slugKey := strings.ToLower(slug)

	if r.isStale() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = r.refreshCache(ctx)
		cancel()
	}

	r.mu.RLock()
	p, found := r.bySlug[slugKey]
	r.mu.RUnlock()

	if found {
		return &p, nil
	}

	// Cache Miss: Query full dataset from Mongo to refresh cache
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.refreshCache(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	p, found = r.bySlug[slugKey]
	r.mu.RUnlock()

	if !found {
		return nil, ErrProjectNotFound
	}
	return &p, nil
}
