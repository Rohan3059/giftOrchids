package controllers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kravi0/BizGrowth-backend/database"
	"github.com/kravi0/BizGrowth-backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var SupportTickerCollection = database.ProductData(database.Client, "CustomerSupportTicket")

func GenerateUniqueTicketID(ctx context.Context,  suffix string) (string, error) {
    // Generate the ticket ID based on creation timestamp and name.
	currentTime,_ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
    generatedTicketID := strconv.FormatInt(currentTime.Unix(), 10) + strings.ToUpper(suffix)

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
	
		ticket.Ticket_id =  primitive.NewObjectID()
		ticket.TicketID = strings.ToUpper(ticket.Ticket_id.String()[0:6])
		ticket.Email = c.PostForm("email")
		ticket.Subject = c.PostForm("subject")
		ticket.Message = c.PostForm("message")
		ticket.Name = c.PostForm("name")
		ticket.MobileNo = c.PostForm("mobileno")
		ticket.Status = "Initiated"
		id,err := GenerateUniqueTicketID(ctx, ticket.Name)
		if(err!=nil){
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to create ticket"})
			return
		}
		ticket.TicketID=id

		ticket.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		 //var multipartForm *multipart.Form
		 multipartForm,err := c.MultipartForm()
		 if(err != nil){
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

			uploadUrl,err := saveFile(f, file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
				return
			}
			attachmentsUrl = append(attachmentsUrl,uploadUrl)

		}
		
		if(ticket.Email == "" || ticket.Subject == "" || ticket.Message == "" || ticket.Name == "" || ticket.MobileNo == ""){

			c.JSON(http.StatusBadRequest, gin.H{"Error": "All fields are required"})
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
		status := c.Query("status")
		defer cancel()
		var tickets []models.CustomerSupportTicket
		
		cursor, err := SupportTickerCollection.Find(ctx, bson.M{})
		if(status != ""){
			cursor, err = SupportTickerCollection.Find(ctx, bson.M{"status": status})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
				return
			}


		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	
		err = cursor.All(ctx, &tickets)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tickets)
	}
}


// @Summary Update ticket status
func UpdateTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ticketId := c.Param("id")

		updatedAt,err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}


		filter := bson.M{"_id": ticketId}

		update := bson.M{"$set": bson.M{
			"status": c.PostForm("status"),
			"updated_at": updatedAt,
		} }

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

func getTicketById() gin.HandlerFunc {
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
		c.JSON(http.StatusOK, ticket)
	}
}


//assign ticket

func assignTicket() gin.HandlerFunc{
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
		updatedAt,err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		filter := bson.M{"_id": ticketId}

		update := bson.M{"$set": bson.M{
			"assigned_to": ticket.AssignedSupport,
			"updated_at": updatedAt,
		} }

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


//add reply to ticket
func addReply() gin.HandlerFunc{
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
		message := c.PostForm("support_reply")
		updatedAt,err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		filter := bson.M{"_id": ticketId}

		update := bson.M{"$push": bson.M{
			"supportmessage": message,
		}, "$set": bson.M{
			"updated_at": updatedAt,
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
