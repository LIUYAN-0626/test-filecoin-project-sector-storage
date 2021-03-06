package sectorstorage

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/filecoin-project/specs-actors/actors/abi"

	"github.com/LIUYAN-0626/test-filecoin-project-sector-storage/sealtasks"
	"github.com/LIUYAN-0626/test-filecoin-project-sector-storage/stores"
)

type existingSelector struct {
	best []stores.SectorStorageInfo
}

func newExistingSelector(ctx context.Context, index stores.SectorIndex, sector abi.SectorID, alloc stores.SectorFileType, allowFetch bool) (*existingSelector, error) {
	best, err := index.StorageFindSector(ctx, sector, alloc, allowFetch)
	if err != nil {
		return nil, err
	}

	return &existingSelector{
		best: best,
	}, nil
}

func (s *existingSelector) Ok(ctx context.Context, task sealtasks.TaskType, spt abi.RegisteredSealProof, whnd *workerHandle) (bool, error) {
	tasks, err := whnd.w.TaskTypes(ctx)
	if err != nil {
		return false, xerrors.Errorf("getting supported worker task types: %w", err)
	}
	if _, supported := tasks[task]; !supported {
		return false, nil
	}

	paths, err := whnd.w.Paths(ctx)
	if err != nil {
		return false, xerrors.Errorf("getting worker paths: %w", err)
	}

	have := map[stores.ID]struct{}{}
	for _, path := range paths {
		have[path.ID] = struct{}{}
	}

	for _, info := range s.best {
		if _, ok := have[info.ID]; ok {
			return true, nil
		}
	}

	return false, nil
}

func (s *existingSelector) Cmp(ctx context.Context, task sealtasks.TaskType, a, b *workerHandle) (bool, error) {
	return a.active.utilization(a.info.Resources) < b.active.utilization(b.info.Resources), nil
}

var _ WorkerSelector = &existingSelector{}
