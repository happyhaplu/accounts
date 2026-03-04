package handlers

import (
"strings"

"outcraftly/accounts/database"
"outcraftly/accounts/models"

"github.com/gofiber/fiber/v2"
)

type profileRequest struct {
Name             string `json:"name"`
CompanyName      string `json:"company_name"`
JobTitle         string `json:"job_title"`
PhoneCountryCode string `json:"phone_country_code"`
PhoneNumber      string `json:"phone_number"`
}

// UpdateProfile sets the user's name, company, and job title.
// Name and CompanyName are required; JobTitle is optional.
func UpdateProfile(c *fiber.Ctx) error {
userID := c.Locals("userID").(string)

req := new(profileRequest)
if err := c.BodyParser(req); err != nil {
return badRequest(c, "Invalid request body")
}

req.Name = strings.TrimSpace(req.Name)
req.CompanyName = strings.TrimSpace(req.CompanyName)
req.JobTitle = strings.TrimSpace(req.JobTitle)
req.PhoneCountryCode = strings.TrimSpace(req.PhoneCountryCode)
req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)

if req.Name == "" {
return badRequest(c, "Name is required")
}
if req.CompanyName == "" {
return badRequest(c, "Company name is required")
}

var user models.User
if tx := database.DB.Where("id = ?", userID).First(&user); tx.Error != nil {
return c.Status(fiber.StatusNotFound).JSON(errJSON("User not found"))
}

database.DB.Model(&user).Updates(map[string]interface{}{
"name":               req.Name,
"company_name":       req.CompanyName,
"job_title":          req.JobTitle,
"phone_country_code": req.PhoneCountryCode,
"phone_number":       req.PhoneNumber,
"profile_complete":   true,
})
user.Name = req.Name
user.CompanyName = req.CompanyName
user.JobTitle = req.JobTitle
user.PhoneCountryCode = req.PhoneCountryCode
user.PhoneNumber = req.PhoneNumber
user.ProfileComplete = true

return c.JSON(fiber.Map{
"message": "Profile saved",
"user":    publicUser(user),
})
}
