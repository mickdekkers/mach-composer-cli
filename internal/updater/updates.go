package updater

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/mach-composer/mach-composer-cli/internal/config"
	"github.com/mach-composer/mach-composer-cli/internal/gitutils"
)

func findUpdates(ctx context.Context, cfg *PartialConfig, filename string) (*UpdateSet, error) {
	log.Ctx(ctx).Info().Msgf("Checking if there are updates for %d components\n", len(cfg.Components))
	if cfg.client == nil {
		return findUpdatesParallel(ctx, cfg, filename)
	}
	return findUpdatesSerial(ctx, cfg, filename)
}

func findUpdatesSerial(ctx context.Context, cfg *PartialConfig, filename string) (*UpdateSet, error) {
	updates := UpdateSet{
		filename: filename,
	}

	for i := range cfg.Components {
		cs, err := getLastVersion(ctx, cfg, &cfg.Components[i], cfg.filename)
		if err != nil {
			return nil, err
		}

		if cs == nil {
			continue
		}

		output := OutputChanges(cs)
		log.Ctx(ctx).Info().Msg(output)

		if cs.HasChanges() {
			updates.updates = append(updates.updates, *cs)
		}
	}
	return &updates, nil
}

func findUpdatesParallel(ctx context.Context, cfg *PartialConfig, filename string) (*UpdateSet, error) {
	numUpdates := len(cfg.Components)
	jobChan := make(chan WorkerJob, numUpdates)
	resChan := make(chan *ChangeSet, numUpdates)
	errChan := make(chan error, numUpdates)

	var wg sync.WaitGroup
	var numWorkers = 4

	// Start 4 workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := range jobChan {
				c := j.component

				logger := log.Ctx(ctx).With().Str("component", c.Name).Logger()

				ctx = logger.WithContext(ctx)

				cs, err := getLastVersion(ctx, cfg, c, cfg.filename)
				if err != nil {
					log.Ctx(ctx).Error().Msg(err.Error())
					errChan <- err
					continue
				}

				if cs == nil {
					continue
				}

				resChan <- cs
				continue
			}
		}()
	}

	// Send work
	for _, c := range cfg.Components {
		component := c
		jobChan <- WorkerJob{
			component: &component,
			cfg:       cfg,
		}
	}
	close(jobChan)

	wg.Wait()
	close(errChan)
	close(resChan)

	if n := len(errChan); n > 0 {
		return nil, fmt.Errorf("failed to update %d components", n)
	}

	// Process results as we receive them from the channel
	updates := UpdateSet{
		filename: filename,
	}
	for changeSet := range resChan {
		if changeSet == nil {
			continue
		}

		output := OutputChanges(changeSet)
		log.Ctx(ctx).Info().Msg(output)

		if changeSet.HasChanges() {
			updates.updates = append(updates.updates, *changeSet)
		}
	}

	return &updates, nil
}

func findSpecificUpdate(ctx context.Context, cfg *PartialConfig, filename string, component *config.Component) (*UpdateSet, error) {
	changeSet, err := getLastVersion(ctx, cfg, component, filename)
	if err != nil {
		return nil, err
	}

	output := OutputChanges(changeSet)
	log.Ctx(ctx).Info().Msg(output)

	updates := UpdateSet{
		filename: cfg.filename,
		updates:  []ChangeSet{*changeSet},
	}
	return &updates, nil
}

func getLastVersion(ctx context.Context, cfg *PartialConfig, c *config.Component, origin string) (*ChangeSet, error) {
	if c.Branch == "" {
		c.Branch = "main"
	}

	if cfg.client != nil {
		return getLastVersionCloud(ctx, cfg, c, origin)
	}

	if strings.HasPrefix(c.Source, "git:") {
		return getLastVersionGit(ctx, c, origin)
	}

	err := &UpdateError{
		msg:       fmt.Sprintf("unrecognized component source for %s: %s", c.Name, c.Source),
		component: c.Name,
		source:    c.Source,
	}
	return nil, err
}

func getLastVersionCloud(ctx context.Context, cfg *PartialConfig, c *config.Component, origin string) (*ChangeSet, error) {
	organization := cfg.MachComposer.Cloud.Organization
	project := cfg.MachComposer.Cloud.Project

	version, _, err := cfg.client.
		ComponentsApi.ComponentLatestVersion(ctx, organization, project, c.Name).
		Branch(c.Branch).
		Execute()

	if err != nil {
		if strings.HasPrefix(c.Source, "git:") {
			log.Ctx(ctx).Warn().Msgf("Error checking for %s in MACH Composer Cloud, falling back to Git", c.Name)
			return getLastVersionGit(ctx, c, origin)
		}
		log.Ctx(ctx).Error().Err(err).Msgf("Error checking for latest version of %s", c.Name)
		return nil, nil
	}

	if version == nil {
		if strings.HasPrefix(c.Source, "git:") {
			log.Ctx(ctx).Warn().Msgf("No version found for %s in MACH Composer Cloud, falling back to Git", c.Name)
			return getLastVersionGit(ctx, c, origin)
		}
		log.Ctx(ctx).Warn().Msgf("No version found for %s", c.Name)
		return nil, nil
	}

	cs := &ChangeSet{
		Changes:     []CommitData{},
		Component:   c,
		LastVersion: version.Version,
	}

	if c.Version != version.Version {
		paginator, _, err := cfg.client.
			ComponentsApi.
			ComponentVersionQueryCommits(ctx, organization, project, c.Name, version.Version).
			Offset(0).
			Limit(200).
			Execute()
		if err != nil {
			return nil, err
		}

		for _, record := range paginator.Results {
			change := CommitData{
				Commit:  record.Commit,
				Parents: record.Parents,
				Message: record.Subject,
				Author: CommitAuthor{
					Email: record.Author.Email,
					Name:  record.Author.Name,
					Date:  record.Author.Date,
				},
				Committer: CommitAuthor{
					Email: record.Committer.Email,
					Name:  record.Committer.Name,
					Date:  record.Committer.Date,
				},
			}
			cs.Changes = append(cs.Changes, change)
		}
	}

	return cs, nil
}

func getLastVersionGit(ctx context.Context, c *config.Component, origin string) (*ChangeSet, error) {
	commits, err := gitutils.GetLastVersionGit(ctx, c, origin)
	if err != nil {
		return nil, err
	}

	cd := make([]CommitData, len(commits))
	for i := range commits {
		c := commits[i]

		cd[i].Commit = c.Commit
		cd[i].Parents = c.Parents
		cd[i].Message = c.Message

		cd[i].Author = CommitAuthor{
			Email: c.Author.Email,
			Name:  c.Author.Name,
			Date:  c.Author.Date,
		}
		cd[i].Committer = CommitAuthor{
			Email: c.Committer.Email,
			Name:  c.Committer.Name,
			Date:  c.Committer.Date,
		}
	}

	cs := &ChangeSet{
		Changes:   cd,
		Component: c,
	}

	if len(commits) < 1 {
		cs.LastVersion = c.Version
	} else {
		cs.LastVersion = commits[0].Commit
	}

	return cs, nil
}
