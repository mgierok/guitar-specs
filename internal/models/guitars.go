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
	Features  []GuitarFeatureResolved // Features for this guitar
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

// GuitarFeatureResolved represents a resolved feature value for display.
type GuitarFeatureResolved struct {
	FeatureKey      string
	FeatureLabel    string
	FeatureKind     string
	ValueDisplay    *string
	EnumValue       *string
	EnumDescription *string
	ValueText       *string
	ValueNumber     *float64
	ValueBoolean    *bool
	Unit            *string
}

// GetBySlug returns a single guitar by slug with brand and shape names.
func (s GuitarStore) GetBySlug(ctx context.Context, slug string) (*Guitar, error) {
	if s.DB == nil {
		return nil, errors.New("nil DB")
	}
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
		where g.slug = $1
	`
	var g Guitar
	if err := s.DB.QueryRow(ctx, q, slug).Scan(
		&g.ID, &g.Slug, &g.Type, &g.Model, &g.BrandSlug, &g.BrandName, &g.ShapeSlug, &g.ShapeName,
	); err != nil {
		return nil, err
	}
	return &g, nil
}

// ListFeaturesBySlug returns resolved features for a guitar identified by slug.
func (s GuitarStore) ListFeaturesBySlug(ctx context.Context, slug string) ([]GuitarFeatureResolved, error) {
	if s.DB == nil {
		return nil, errors.New("nil DB")
	}
	var cancel func()
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}
	const fq = `
SELECT
  f.key          AS feature_key,
  f.label        AS feature_label,
  f.kind::text   AS feature_kind,
  COALESCE(
    fav.value,
    gf.value_text,
    CASE WHEN gf.value_number IS NOT NULL
      THEN (gf.value_number::text || COALESCE(' '||f.unit, '')) END,
    CASE WHEN gf.value_boolean IS NOT NULL
      THEN CASE WHEN gf.value_boolean THEN 'true' ELSE 'false' END END
  )                AS value_display,
  fav.value        AS enum_value,
  fav.description  AS enum_description,
  gf.value_text,
  gf.value_number::float8,
  gf.value_boolean,
  f.unit
FROM public.guitars g
JOIN public.guitar_features gf         ON gf.guitar_id = g.id
JOIN public.features f                 ON f.id = gf.feature_id
LEFT JOIN public.feature_allowed_values fav ON fav.id = gf.allowed_value_id
WHERE g.slug = $1
ORDER BY f.label;
	`
	rows, err := s.DB.Query(ctx, fq, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]GuitarFeatureResolved, 0, 32)
	for rows.Next() {
		var r GuitarFeatureResolved
		if err := rows.Scan(
			&r.FeatureKey,
			&r.FeatureLabel,
			&r.FeatureKind,
			&r.ValueDisplay,
			&r.EnumValue,
			&r.EnumDescription,
			&r.ValueText,
			&r.ValueNumber,
			&r.ValueBoolean,
			&r.Unit,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
