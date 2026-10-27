package repository

import (
	"context"
	"errors"

	"api-students/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrPortfolioNotFound    = errors.New("portofolio tidak ditemukan")
	ErrProjectLimitExceeded = errors.New("jumlah proyek sudah mencapai batas maksimal (20 proyek)")
	ErrProjectNotFound     = errors.New("proyek tidak ditemukan")
)

type PortfolioRepository interface {
	GetByStudentID(ctx context.Context, studentID int64) (*model.Portfolio, error)
	UpsertPortfolio(ctx context.Context, portfolio *model.Portfolio) error
	AddProject(ctx context.Context, studentID int64, project *model.Project) error
	DeleteProject(ctx context.Context, studentID int64, projectID string) error
	GetTopTechnologies(ctx context.Context) ([]model.TechStat, error)
}

type mongoPortfolioRepository struct {
	collection *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &mongoPortfolioRepository{
		collection: db.Collection("student_portfolios"),
	}
}

// Struct internal BSON khusus layer repository
type mongoProject struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Judul     string             `bson:"judul"`
	Teknologi []string           `bson:"teknologi"`
	Tautan    string             `bson:"tautan"`
	Tahun     int                `bson:"tahun"`
}

type mongoCertificate struct {
	Nama     string `bson:"nama"`
	Penerbit string `bson:"penerbit"`
	Tanggal  string `bson:"tanggal"`
}

type mongoPortfolio struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	StudentID    int64              `bson:"student_id"`
	Headline     string             `bson:"headline"`
	Bio          string             `bson:"bio"`
	Projects     []mongoProject     `bson:"projects"`
	Certificates []mongoCertificate `bson:"certificates"`
}

func (r *mongoPortfolioRepository) GetByStudentID(ctx context.Context, studentID int64) (*model.Portfolio, error) {
	var doc mongoPortfolio
	filter := bson.M{"student_id": studentID}

	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPortfolioNotFound
		}
		return nil, err
	}

	return toDomainPortfolio(&doc), nil
}

func (r *mongoPortfolioRepository) UpsertPortfolio(ctx context.Context, p *model.Portfolio) error {
	filter := bson.M{"student_id": p.StudentID}

	mongoProjects := make([]mongoProject, len(p.Projects))
	for i, prj := range p.Projects {
		var objID primitive.ObjectID
		if prj.ID != "" {
			objID, _ = primitive.ObjectIDFromHex(prj.ID)
		} else {
			objID = primitive.NewObjectID()
		}
		mongoProjects[i] = mongoProject{
			ID:        objID,
			Judul:     prj.Judul,
			Teknologi: prj.Teknologi,
			Tautan:    prj.Tautan,
			Tahun:     prj.Tahun,
		}
	}

	mongoCerts := make([]mongoCertificate, len(p.Certificates))
	for i, cert := range p.Certificates {
		mongoCerts[i] = mongoCertificate{
			Nama:     cert.Nama,
			Penerbit: cert.Penerbit,
			Tanggal:  cert.Tanggal,
		}
	}

	update := bson.M{
		"$set": bson.M{
			"student_id":   p.StudentID,
			"headline":     p.Headline,
			"bio":          p.Bio,
			"projects":     mongoProjects,
			"certificates": mongoCerts,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *mongoPortfolioRepository) AddProject(ctx context.Context, studentID int64, prj *model.Project) error {
	// 1. Cek apakah dokumen portofolio mahasiswa sudah ada
	count, err := r.collection.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrPortfolioNotFound
	}

	// 2. Pola atomik "projects.19": { "$exists": false }
	// Jika index ke-19 ada, artinya proyek sudah berjumlah 20 (penambahan ditolak)
	filter := bson.M{
		"student_id":  studentID,
		"projects.19": bson.M{"$exists": false},
	}

	newID := primitive.NewObjectID()
	prj.ID = newID.Hex()

	newMongoPrj := mongoProject{
		ID:        newID,
		Judul:     prj.Judul,
		Teknologi: prj.Teknologi,
		Tautan:    prj.Tautan,
		Tahun:     prj.Tahun,
	}

	update := bson.M{
		"$push": bson.M{
			"projects": newMongoPrj,
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return ErrProjectLimitExceeded
	}

	return nil
}

func (r *mongoPortfolioRepository) DeleteProject(ctx context.Context, studentID int64, projectID string) error {
	objID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return ErrProjectNotFound
	}

	filter := bson.M{"student_id": studentID}
	update := bson.M{
		"$pull": bson.M{
			"projects": bson.M{"_id": objID},
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return ErrPortfolioNotFound
	}
	if res.ModifiedCount == 0 {
		return ErrProjectNotFound
	}

	return nil
}

func (r *mongoPortfolioRepository) GetTopTechnologies(ctx context.Context) ([]model.TechStat, error) {
	// Aggregation pipeline dengan dua kali $unwind (sesuai petunjuk modul)
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$unwind", Value: "$projects"}},
		bson.D{{Key: "$unwind", Value: "$projects.teknologi"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$projects.teknologi"},
			{Key: "jumlah", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "jumlah", Value: -1}}}},
		bson.D{{Key: "$limit", Value: 10}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID     string `bson:"_id"`
		Jumlah int    `bson:"jumlah"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	stats := make([]model.TechStat, len(results))
	for i, res := range results {
		stats[i] = model.TechStat{
			Teknologi: res.ID,
			Jumlah:    res.Jumlah,
		}
	}

	return stats, nil
}

func toDomainPortfolio(doc *mongoPortfolio) *model.Portfolio {
	projects := make([]model.Project, len(doc.Projects))
	for i, prj := range doc.Projects {
		projects[i] = model.Project{
			ID:        prj.ID.Hex(),
			Judul:     prj.Judul,
			Teknologi: prj.Teknologi,
			Tautan:    prj.Tautan,
			Tahun:     prj.Tahun,
		}
	}

	certs := make([]model.Certificate, len(doc.Certificates))
	for i, cert := range doc.Certificates {
		certs[i] = model.Certificate{
			Nama:     cert.Nama,
			Penerbit: cert.Penerbit,
			Tanggal:  cert.Tanggal,
		}
	}

	return &model.Portfolio{
		ID:           doc.ID.Hex(),
		StudentID:    doc.StudentID,
		Headline:     doc.Headline,
		Bio:          doc.Bio,
		Projects:     projects,
		Certificates: certs,
	}
}