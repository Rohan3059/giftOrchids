package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	"github.com/kravi0/BizGrowth-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var SupportTickerCollection = database.ProductData(database.Client, "CustomerSupportTicket")
var SupportChatMessage = database.ProductData(database.Client, "SupportChatMessage")

func GenerateUniqueTicketID(ctx context.Context, suffix string) (string, error) {
	// Generate the ticket ID based on creation timestamp and name.
	currentTime, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	generatedTicketID := strconv.FormatInt(currentTime.Unix(), 5) + strings.ToUpper(suffix)[0:2]

	// Check if the generated ID already exists.
	for {
		count, err := SupportTickerCollection.CountDocuments(ctx, bson.M{"ticketid": generatedTicketID})
		if err != nil {
			return "", err
		}
		if count == 0 {
			break
		}
		// If the ID already exists, append a suffix to make it unique.
		generatedTicketID += strconv.Itoa(int(count))
	}

	return generatedTicketID, nil
}

func CreateTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		var ticket models.CustomerSupportTicket

		//generate random 6 digit ticket id not based on time

		ticket.Ticket_id = primitive.NewObjectID()
		ticket.Email = c.PostForm("email")
		ticket.Subject = c.PostForm("subject")
		ticket.Message = c.PostForm("message")
		ticket.Name = c.PostForm("name")
		ticket.MobileNo = c.PostForm("mobileno")
		ticket.Status = "Initiated"
		id, err := GenerateUniqueTicketID(ctx, ticket.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to create ticket"})
			return
		}
		ticket.TicketID = id

		ticket.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		ticket.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		//var multipartForm *multipart.Form
		multipartForm, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		files := multipartForm.File["attachments"]
		var attachmentsUrl []string
		for _, file := range files {
			f, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
				return
			}
			defer f.Close()

			uploadUrl, err := saveFile(f, file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
				return
			}
			attachmentsUrl = append(attachmentsUrl, uploadUrl)

		}

		if ticket.Email == "" || ticket.Subject == "" || ticket.Message == "" || ticket.Name == "" || ticket.MobileNo == "" {

			c.JSON(http.StatusBadRequest, gin.H{"Error": "All fields are required"})
			return
		}

		ticket.Attachments = attachmentsUrl

		_, insertErr := SupportTickerCollection.InsertOne(ctx, ticket)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": insertErr.Error()})
			return
		}
		c.JSON(http.StatusOK, ticket)
	}

}

// @Summary Get all tickets
func GetTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		status := c.Query("status")
		filter := bson.M{}
		if status != "" {
			filter["status"] = status
		}

		cursor, err := SupportTickerCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var tickets []models.CustomerSupportTicket
		if err := cursor.All(ctx, &tickets); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		//go through each ticket , iterate over attachments, create presignurl and add back

		for i, ticket := range tickets {
			for j, attachment := range ticket.Attachments {
				url, err := getPresignURL(attachment)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
					return
				}
				tickets[i].Attachments[j] = url
			}
		}

		c.JSON(http.StatusOK, tickets)
	}
}

// @Summary Update ticket status
func UpdateTicketStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ticketId := c.Param("id")

		updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		statusRequest := c.PostForm("status")
		var status string

		//check status
		switch statusRequest {
		case "Initiated":
			status = utils.Initiated
		case "initiated":
			status = utils.Initiated
		case "in progress":
			status = utils.InProgress
		case "In Progress":
			status = utils.InProgress
		case "Closed":
			status = utils.Closed
		case "closed":
			status = utils.Closed
		default:
		}

		filter := bson.M{"ticketid": ticketId}

		update := bson.M{"$set": bson.M{
			"status":     status,
			"updated_at": updatedAt,
		}}

		var ticket models.CustomerSupportTicket

		updateErr := SupportTickerCollection.FindOneAndUpdate(ctx, filter, update).Decode(&ticket)
		if updateErr != nil {
			if updateErr == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"Error": "Ticket not found"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Ticket updated successfully",
		})
	}

}

func GetTicketById() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ticketID := c.Param("id")
		fmt.Print(ticketID)
		filter := bson.M{
			"$or": []bson.M{
				{"_id": ticketID},
				{"ticketid": ticketID},
			},
		}

		if ticketID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Ticket ID is required"})
			return
		}

		var ticket models.CustomerSupportTicket
		err := SupportTickerCollection.FindOne(ctx, filter).Decode(&ticket)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		for j, attachment := range ticket.Attachments {
			url, err := getPresignURL(attachment)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
				return
			}
			ticket.Attachments[j] = url
		}

		c.JSON(http.StatusOK, ticket)
	}
}

func AssignTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ticketId := c.Param("id")
		var ticket models.CustomerSupportTicket
		err := SupportTickerCollection.FindOne(ctx, bson.M{"_id": ticketId}).Decode(&ticket)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		ticket.AssignedSupport = c.PostForm("assignedto")
		updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		filter := bson.M{"_id": ticketId}

		update := bson.M{"$set": bson.M{
			"assigned_to": ticket.AssignedSupport,
			"updated_at":  updatedAt,
		}}

		_, updateErr := SupportTickerCollection.UpdateOne(ctx, filter, update)
		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": updateErr.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Ticket updated successfully",
		})
	}
}

func GetTicketCounts() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		type TicketCounts struct {
			Initiated  int
			InProgress int
			Closed     int
		}

		statusMap := map[string]int{
			"Initiated":   1,
			"In Progress": 2,
			"Closed":      3,
		}

		counts := TicketCounts{}

		for status, code := range statusMap {
			count, err := SupportTickerCollection.CountDocuments(ctx, bson.M{"status": status})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			switch code {
			case 1:
				counts.Initiated = int(count)
			case 2:
				counts.InProgress = int(count)
			case 3:
				counts.Closed = int(count)
				// Add more cases for other statuses if needed
			}
		}

		c.JSON(http.StatusOK, counts)
	}
}

func AddMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ticketId := c.Param("id")

		if !checkAdmin(ctx, c) {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "Unauthorized"})
			return
		}

		if ticketId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "ticket id is required"})
			return
		}

		var ticket models.CustomerSupportTicket
		findErr := SupportTickerCollection.FindOne(ctx, bson.M{"ticketid": ticketId}).Decode(&ticket)
		if findErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Invalid support ticket"})
			return
		}

		var message models.ChatMessage
		message.SentAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		message.Sender = "Admin"

		if err := c.BindJSON(&message); err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		var existingChatMessage models.SupportChatMessage

		finderr := SupportChatMessage.FindOne(ctx, bson.M{"ticketId": ticketId}).Decode(&existingChatMessage)
		if finderr != nil {

			//create new
			chatMessage := models.SupportChatMessage{
				SupportChatId:   primitive.NewObjectID(),
				SupportTicketId: ticketId,
				Messages:        []models.ChatMessage{message},
			}

			_, insertErr := SupportChatMessage.InsertOne(ctx, chatMessage)
			if insertErr != nil {
				fmt.Println(insertErr)
				c.JSON(http.StatusBadRequest, gin.H{"Error": insertErr.Error()})
				return
			}
			//update SupportTicker with id of SupportChatMessage

			_, updateErr := SupportTickerCollection.UpdateOne(ctx, bson.M{"_id": ticketId}, bson.M{"$set": bson.M{
				"chatmessage": chatMessage.SupportChatId,
				"updated_at":  updatedAt,
			}})
			if updateErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": updateErr.Error()})
				return
			}
		}

		_, updateErr := SupportChatMessage.UpdateOne(ctx, bson.M{"ticketId": ticketId}, bson.M{"$push": bson.M{
			"messages": message,
		}})
		if updateErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": updateErr.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Message has been sent successfully",
		})

	}

}

func AddSellerMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ticketId := c.Param("id")

		if !checkSeller(ctx, c) {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "Unauthorized"})
			return
		}

		uid, exist := c.Get("uid")

		if !exist {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "Unauthorized"})
			return
		}

		// Convert to seller id objectid
		sellerId, _ := primitive.ObjectIDFromHex(uid.(string))

		var foundSeller models.Seller
		err := SellerCollection.FindOne(ctx, bson.M{"_id": sellerId}).Decode(&foundSeller)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unable to find seller"})
			return
		}

		name := foundSeller.Company_Name

		if ticketId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Ticket id is required"})
			return
		}

		var ticket models.CustomerSupportTicket
		findErr := SupportTickerCollection.FindOne(ctx, bson.M{"ticketid": ticketId}).Decode(&ticket)
		if findErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Invalid support ticket"})
			return
		}

		var message models.ChatMessage
		message.SentAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		message.Sender = name

		if err := c.BindJSON(&message); err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updatedAt := time.Now()

		var existingChatMessage models.SupportChatMessage
		finderr := SupportChatMessage.FindOne(ctx, bson.M{"SupportTicketId": ticketId}).Decode(&existingChatMessage)
		if finderr != nil {
			// Create new chat message document
			chatMessage := models.SupportChatMessage{
				SupportChatId:   primitive.NewObjectID(),
				SupportTicketId: ticketId,
				Messages:        []models.ChatMessage{message},
			}

			_, insertErr := SupportChatMessage.InsertOne(ctx, chatMessage)
			if insertErr != nil {
				fmt.Println(insertErr)
				c.JSON(http.StatusBadRequest, gin.H{"Error": insertErr.Error()})
				return
			}

			// Update SupportTicket with id of SupportChatMessage
			_, updateErr := SupportTickerCollection.UpdateOne(ctx, bson.M{"ticketid": ticketId}, bson.M{"$set": bson.M{
				"chatmessage": chatMessage.SupportChatId,
				"updated_at":  updatedAt,
			}})
			if updateErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": updateErr.Error()})
				return
			}
		} else {
			// Add message to existing chat document
			_, updateErr := SupportChatMessage.UpdateOne(ctx, bson.M{"SupportTicketId": ticketId}, bson.M{"$push": bson.M{
				"messages": message,
			}})
			if updateErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": updateErr.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Message has been sent successfully",
		})
	}
}

func GetChatMessagesHandler() gin.HandlerFunc {

	return func(c *gin.Context) {

		ticketID := c.Param("id")

		if ticketID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "ticket id is required"})
			return
		}

		messages, err := GetChatMessagesByTicketID(ticketID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"messages": messages})
	}
}

func GetChatMessagesByTicketID(ticketID string) ([]models.ChatMessage, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var supportChatMessage models.SupportChatMessage
	err := SupportChatMessage.FindOne(ctx, bson.M{"ticketId": ticketID}).Decode(&supportChatMessage)
	if err != nil {
		return nil, err
	}

	return supportChatMessage.Messages, nil
}
