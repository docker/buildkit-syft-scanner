package filedigest

import (
	"context"
	"crypto"
	"errors"
	"fmt"

	"github.com/dustin/go-humanize"

	"github.com/anchore/go-sync"
	stereoscopeFile "github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/syft/internal"
	"github.com/anchore/syft/internal/bus"
	intFile "github.com/anchore/syft/internal/file"
	"github.com/anchore/syft/internal/log"
	"github.com/anchore/syft/internal/unknown"
	"github.com/anchore/syft/syft/cataloging"
	"github.com/anchore/syft/syft/event/monitor"
	"github.com/anchore/syft/syft/file"
	intCataloger "github.com/anchore/syft/syft/file/cataloger/internal"
)

var ErrUndigestableFile = errors.New("undigestable file")

type Cataloger struct {
	hashes []crypto.Hash
}

func NewCataloger(hashes []crypto.Hash) *Cataloger {
	return &Cataloger{
		hashes: intFile.NormalizeHashes(hashes),
	}
}

func (i *Cataloger) Catalog(ctx context.Context, resolver file.Resolver, coordinates ...file.Coordinates) (map[file.Coordinates][]file.Digest, error) {
	results := make(map[file.Coordinates][]file.Digest)
	var locations []file.Location

	if len(coordinates) == 0 {
		locations = intCataloger.AllRegularFiles(ctx, resolver)
	} else {
		for _, c := range coordinates {
			locs, err := resolver.FilesByPath(c.RealPath)
			if err != nil {
				return nil, fmt.Errorf("unable to get file locations for path %q: %w", c.RealPath, err)
			}
			locations = append(locations, locs...)
		}
	}

	prog := catalogingProgress(int64(len(locations)))

	err := sync.Collect(&ctx, cataloging.ExecutorFile, sync.ToSeq(locations), func(location file.Location) ([]file.Digest, error) {
		result, err := i.catalogLocation(ctx, resolver, location)

		if errors.Is(err, ErrUndigestableFile) {
			return nil, nil
		}

		prog.AtomicStage.Set(location.Path())

		if internal.IsErrPathPermission(err) {
			log.Debugf("file digests cataloger skipping %q: %+v", location.RealPath, err)
			return nil, unknown.New(location, err)
		}

		if err != nil {
			prog.SetError(err)
			return nil, unknown.New(location, err)
		}

		prog.Increment()

		return result, nil
	}, func(location file.Location, digests []file.Digest) {
		if len(digests) > 0 {
			results[location.Coordinates] = digests
		}
	})

	log.Debugf("file digests cataloger processed %d files", prog.Current())

	prog.AtomicStage.Set(fmt.Sprintf("%s files", humanize.Comma(prog.Current())))
	prog.SetCompleted()

	return results, err
}

func (i *Cataloger) catalogLocation(ctx context.Context, resolver file.Resolver, location file.Location) ([]file.Digest, error) {
	meta, err := resolver.FileMetadataByLocation(location)
	if err != nil {
		return nil, err
	}

	// we should only attempt to report digests for files that are regular files (don't attempt to resolve links)
	if meta.Type != stereoscopeFile.TypeRegular {
		return nil, ErrUndigestableFile
	}

	contentReader, err := resolver.FileContentsByLocation(location)
	if err != nil {
		return nil, err
	}
	defer internal.CloseAndLogError(contentReader, location.AccessPath)

	digests, err := intFile.NewDigestsFromFile(ctx, contentReader, i.hashes)
	if err != nil {
		return nil, internal.ErrPath{Context: "digests-cataloger", Path: location.RealPath, Err: err}
	}

	return digests, nil
}

func catalogingProgress(locations int64) *monitor.TaskProgress {
	info := monitor.GenericTask{
		Title: monitor.Title{
			Default: "File digests",
		},
		ParentID: monitor.TopLevelCatalogingTaskID,
	}

	return bus.StartCatalogerTask(info, locations, "")
}
