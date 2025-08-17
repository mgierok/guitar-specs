package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Guitar mirrors selected fields of public.guitars for application usage.
type Guitar struct {
	ID        string
	Slug      string
	Type      string
	Model     string
	BrandSlug string
	BrandName string
	ShapeSlug string
	ShapeName string
}

// GuitarStore provides read operations over guitars.
type GuitarStore struct {
	DB *pgxpool.Pool
}

// List returns guitars ordered by brand, model. Context has a safety timeout.
func (s GuitarStore) List(ctx context.Context) ([]Guitar, error) {
	if s.DB == nil {
		return nil, errors.New("nil DB")
	}

	// Apply a short safety timeout to avoid lingering queries if caller forgot one.
	var cancel func()
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	const q = `
		select 
			g.id::text,
			g.slug::text,
			g.type::text,
			g.model,
			b.slug::text as brand_slug,
			b.name        as brand_name,
			s.slug::text  as shape_slug,
			s.name        as shape_name
		from public.guitars g
		join public.brands b on b.slug = g.brand_slug
		join public.shapes s on s.slug = g.shape_slug
		order by b.name, g.model
	`
	rows, err := s.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	guitars := make([]Guitar, 0, 64)
	for rows.Next() {
		var g Guitar
		if err := rows.Scan(
			&g.ID,
			&g.Slug,
			&g.Type,
			&g.Model,
			&g.BrandSlug,
			&g.BrandName,
			&g.ShapeSlug,
			&g.ShapeName,
		); err != nil {
			return nil, err
		}
		guitars = append(guitars, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return guitars, nil
}
