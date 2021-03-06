package api

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/stashapp/stash/pkg/database"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/utils"
)

func (r *mutationResolver) StudioCreate(ctx context.Context, input models.StudioCreateInput) (*models.Studio, error) {
	// generate checksum from studio name rather than image
	checksum := utils.MD5FromString(input.Name)

	var imageData []byte
	var err error

	if input.Image == nil {
		input.Image = &models.DefaultStudioImage
	}

	// Process the base 64 encoded image string
	_, imageData, err = utils.ProcessBase64Image(*input.Image)
	if err != nil {
		return nil, err
	}

	// Populate a new studio from the input
	currentTime := time.Now()
	newStudio := models.Studio{
		Image:     imageData,
		Checksum:  checksum,
		Name:      sql.NullString{String: input.Name, Valid: true},
		CreatedAt: models.SQLiteTimestamp{Timestamp: currentTime},
		UpdatedAt: models.SQLiteTimestamp{Timestamp: currentTime},
	}
	if input.URL != nil {
		newStudio.URL = sql.NullString{String: *input.URL, Valid: true}
	}

	// Start the transaction and save the studio
	tx := database.DB.MustBeginTx(ctx, nil)
	qb := models.NewStudioQueryBuilder()
	studio, err := qb.Create(newStudio, tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return studio, nil
}

func (r *mutationResolver) StudioUpdate(ctx context.Context, input models.StudioUpdateInput) (*models.Studio, error) {
	// Populate studio from the input
	studioID, _ := strconv.Atoi(input.ID)
	updatedStudio := models.Studio{
		ID:        studioID,
		UpdatedAt: models.SQLiteTimestamp{Timestamp: time.Now()},
	}
	if input.Image != nil {
		_, imageData, err := utils.ProcessBase64Image(*input.Image)
		if err != nil {
			return nil, err
		}
		updatedStudio.Image = imageData
	}
	if input.Name != nil {
		// generate checksum from studio name rather than image
		checksum := utils.MD5FromString(*input.Name)
		updatedStudio.Name = sql.NullString{String: *input.Name, Valid: true}
		updatedStudio.Checksum = checksum
	}
	if input.URL != nil {
		updatedStudio.URL = sql.NullString{String: *input.URL, Valid: true}
	}

	// Start the transaction and save the studio
	tx := database.DB.MustBeginTx(ctx, nil)
	qb := models.NewStudioQueryBuilder()
	studio, err := qb.Update(updatedStudio, tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return studio, nil
}

func (r *mutationResolver) StudioDestroy(ctx context.Context, input models.StudioDestroyInput) (bool, error) {
	qb := models.NewStudioQueryBuilder()
	tx := database.DB.MustBeginTx(ctx, nil)
	if err := qb.Destroy(input.ID, tx); err != nil {
		_ = tx.Rollback()
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
