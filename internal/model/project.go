package model

import "github.com/google/uuid"

type Project struct {
	ID        uuid.UUID
	Name      string
	OwnerID   uuid.UUID
	OwnerName string
}

func NewProject(id uuid.UUID, name string, ownerID uuid.UUID, ownerName string) *Project {
	return &Project{ID: id, Name: name, OwnerID: ownerID, OwnerName: ownerName}
}

func (p *Project) ToTempl() *ProjectTempl {
	return &ProjectTempl{
		ID:        p.ID.String(),
		Name:      p.Name,
		OwnerID:   p.OwnerID.String(),
		OwnerName: p.OwnerName,
	}
}

func (p *Project) ToDB() *ProjectDB {
	return &ProjectDB{
		ID:        p.ID.String(),
		Name:      p.Name,
		OwnerID:   p.OwnerID.String(),
		OwnerName: p.OwnerName,
	}
}

type ProjectTempl struct {
	ID        string
	Name      string
	OwnerID   string
	OwnerName string
}

type ProjectDB struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	OwnerID   string `db:"owner_id"`
	OwnerName string `db:"owner_name"`
}

func ProjectFromDB(p *ProjectDB) (*Project, error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, err
	}

	ownerID, err := uuid.Parse(p.OwnerID)
	if err != nil {
		return nil, err
	}

	return &Project{
		ID:        id,
		Name:      p.Name,
		OwnerID:   ownerID,
		OwnerName: p.OwnerName,
	}, nil
}
