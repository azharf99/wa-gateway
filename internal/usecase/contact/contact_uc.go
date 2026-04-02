package contact

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"

	"github.com/azharf99/wa-gateway/internal/domain"
	// ...
)

type contactUsecase struct {
	repo domain.ContactRepository
}

func NewContactUsecase(repo domain.ContactRepository) domain.ContactUsecase {
	return &contactUsecase{
		repo: repo,
	}
}

func (uc *contactUsecase) ListContacts(ctx context.Context) ([]domain.Contact, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *contactUsecase) AddContact(ctx context.Context, c domain.Contact) error {
	return uc.repo.Create(ctx, &c)
}

func (uc *contactUsecase) UpdateContact(ctx context.Context, c domain.Contact) error {
	return uc.repo.Update(ctx, &c)
}

func (uc *contactUsecase) RemoveContact(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *contactUsecase) ImportFromCSV(ctx context.Context, fileBytes []byte) error {
	reader := csv.NewReader(bytes.NewReader(fileBytes))

	// Lewati baris pertama (header)
	_, err := reader.Read()
	if err != nil {
		return err
	}

	var contacts []domain.Contact
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Pastikan record memiliki minimal 2 kolom (Nama & Phone)
		if len(record) >= 2 {
			contacts = append(contacts, domain.Contact{
				Name:     record[0],
				Phone:    record[1],
				Category: record[2], // Opsional
			})
		}
	}

	return uc.repo.ImportCSV(ctx, contacts)
}
