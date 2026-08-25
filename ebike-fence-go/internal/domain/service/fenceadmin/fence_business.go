package fenceadmin

import (
	"context"

	"ebike-fence-go/internal/api/dto"
)

type FenceBusinessAdmin struct{}

func NewFenceBusinessAdmin() *FenceBusinessAdmin { return &FenceBusinessAdmin{} }

func (f *FenceBusinessAdmin) QueryByFenceIDs(ctx context.Context, ids []int64) ([]dto.FenceCO, error) {
	r, err := requireFenceRepo()
	if err != nil {
		return nil, err
	}
	m, err := r.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]dto.FenceCO, 0, len(ids))
	for _, id := range ids {
		if fe, ok := m[id]; ok {
			out = append(out, toFenceCO(fe))
		}
	}
	return out, nil
}

func (f *FenceBusinessAdmin) QueryByFenceID(ctx context.Context, id int64) (*dto.FenceCO, error) {
	fe, err := getFenceOrErr(ctx, id)
	if err != nil {
		return nil, err
	}
	co := toFenceCO(*fe)
	return &co, nil
}
