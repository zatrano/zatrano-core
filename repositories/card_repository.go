package repositories

import (
	"context"

	"zatrano/configs/databaseconfig"
	"zatrano/models"
	"zatrano/pkg/queryparams"

	"gorm.io/gorm"
)

type ICardRepository interface {
	GetAllCards(params queryparams.ListParams) ([]models.Card, int64, error)
	GetCardByID(id uint) (*models.Card, error)
	CreateCardWithRelations(ctx context.Context, card *models.Card) error
	UpdateCardWithRelations(ctx context.Context, card *models.Card) error
	DeleteCardWithRelations(ctx context.Context, id uint) error
	GetCardCount() (int64, error)
	IsSlugAvailable(slug string, excludeID uint) (bool, error)
}

type CardRepository struct {
	base IBaseRepository[models.Card]
	db   *gorm.DB
}

func NewCardRepository() ICardRepository {
	base := NewBaseRepository[models.Card](databaseconfig.GetDB())
	base.SetAllowedSortColumns([]string{"id", "name", "slug", "is_active", "created_at"})
	base.SetPreloads("CardBanks.Bank", "CardSocialMedia.SocialMedia")
	return &CardRepository{base: base, db: databaseconfig.GetDB()}
}

func (r *CardRepository) GetAllCards(params queryparams.ListParams) ([]models.Card, int64, error) {
	return r.base.GetAll(params)
}

func (r *CardRepository) GetCardByID(id uint) (*models.Card, error) {
	return r.base.GetByID(id)
}

func (r *CardRepository) CreateCardWithRelations(ctx context.Context, card *models.Card) error {
	return r.base.CreateWithRelations(ctx, card)
}

func (r *CardRepository) UpdateCardWithRelations(ctx context.Context, card *models.Card) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.Card{}).Where("id = ?", card.ID).Omit("CardBanks", "CardSocialMedia").Updates(card).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("card_id = ?", card.ID).Delete(&models.CardBank{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("card_id = ?", card.ID).Delete(&models.CardSocialMedia{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(card.CardBanks) > 0 {
		for i := range card.CardBanks {
			card.CardBanks[i].CardID = card.ID
		}
		if err := tx.Create(&card.CardBanks).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(card.CardSocialMedia) > 0 {
		for i := range card.CardSocialMedia {
			card.CardSocialMedia[i].CardID = card.ID
		}
		if err := tx.Create(&card.CardSocialMedia).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *CardRepository) DeleteCardWithRelations(ctx context.Context, id uint) error {
	return r.base.DeleteWithRelations(ctx, id)
}

func (r *CardRepository) GetCardCount() (int64, error) {
	return r.base.GetCount()
}

func (r *CardRepository) IsSlugAvailable(slug string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Unscoped().Model(&models.Card{}).Where("slug = ?", slug)

	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
