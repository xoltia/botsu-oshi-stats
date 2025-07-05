package index

import (
	"context"

	"github.com/xoltia/botsu-oshi-stats/logs"
	"github.com/xoltia/botsu-oshi-stats/vtubers"
)

// Indexer reads from the log repository and populates
// relevant information based on the last update to the
// vtuber store.
type Indexer struct {
	vtuberStore *vtubers.Store
	logRepo     *logs.UserLogRepository
	indexRepo   *IndexedVideoRepository
}

func NewIndexer(
	vs *vtubers.Store,
	lr *logs.UserLogRepository,
	ir *IndexedVideoRepository,
) *Indexer {
	return &Indexer{vs, lr, ir}
}

// TODO: partial reindexing when data not recently updated
func (i *Indexer) Index(ctx context.Context) error {
	detector, err := vtubers.CreateDetector(ctx, i.vtuberStore)
	if err != nil {
		return err
	}

	ls, err := i.logRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	defer ls.Close()

	for ls.Next() {
		log, err := ls.Scan()
		if err != nil {
			return err
		}

		vtubers, err := detector.Detect(ctx, log)
		if err != nil {
			return err
		}

		for _, v := range vtubers.All {
			err := i.indexRepo.InsertVideoVTuber(
				ctx,
				log.UserID,
				log.Video.ID,
				v.ID)
			if err != nil {
				return err
			}
		}

		err = i.indexRepo.InsertVideoHistory(
			ctx,
			log.UserID,
			log.Video.ID,
			log.Date,
			log.Duration)
		if err != nil {
			return err
		}
	}

	if err := ls.Err(); err != nil {
		return err
	}

	return nil
}
