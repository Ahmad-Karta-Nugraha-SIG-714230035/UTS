package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GISFeature struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	Lat      float64            `json:"lat" bson:"lat"`
	Lng      float64            `json:"lng" bson:"lng"`
	Category string             `json:"category" bson:"category"`
}

var client *mongo.Client
var collection *mongo.Collection

func initMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Printf("Warning: Could not connect to MongoDB: %v. Running without database.", err)
		return
	}
	collection = client.Database("gisdb").Collection("features")
	log.Println("Connected to MongoDB successfully")
}

func getFeatures(c *fiber.Ctx) error {
	if collection == nil {
		return c.JSON([]GISFeature{}) // Return empty array if no DB
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	var features []GISFeature
	if err = cursor.All(ctx, &features); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(features)
}

func addFeature(c *fiber.Ctx) error {
	if collection == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not connected"})
	}
	var feature GISFeature
	if err := c.BodyParser(&feature); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, feature)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(feature)
}

func updateFeature(c *fiber.Ctx) error {
	if collection == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not connected"})
	}
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var feature GISFeature
	if err := c.BodyParser(&feature); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"name": feature.Name, "lat": feature.Lat, "lng": feature.Lng, "category": feature.Category}}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Feature updated successfully"})
}

func deleteFeature(c *fiber.Ctx) error {
	if collection == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database not connected"})
	}
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"_id": objectID}

	_, err = collection.DeleteOne(ctx, filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Feature deleted successfully"})
}

func main() {
	initMongo()
	app := fiber.New()

	// CORS for frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/api/features", getFeatures)
	app.Post("/api/features", addFeature)
	app.Put("/api/features/:id", updateFeature)
	app.Delete("/api/features/:id", deleteFeature)

	// Serve static files from root
	app.Static("/", "./")

	// Route for locations management page
	app.Get("/locations", func(c *fiber.Ctx) error {
		return c.SendFile("locations.html")
	})

	log.Fatal(app.Listen(":8080"))
}
