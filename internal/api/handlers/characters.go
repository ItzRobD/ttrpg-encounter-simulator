package handlers

// TODO: We're working on saving custom characters, spellcasting is broken
// 	Need to fix the frontend / compare what the backend expects

//func GetCharacterSummaries(c *gin.Context) {
//	summaries, err := character.GetCustomCharacterSummaries(c.Request.Context())
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get character summaries: " + err.Error()})
//		log.Printf("GetCharacterSummaries: Failed to get character summaries: %v", err)
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{
//		"data":  summaries,
//		"count": len(summaries),
//	})
//}
//
//func GetCharacterByID(c *gin.Context) {
//	idStr := c.Param("id")
//	if idStr == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "character ID is required"})
//		return
//	}
//
//	char, err := character.GetCustomCharacterByID(c.Request.Context(), idStr)
//	if err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
//		log.Printf("GetCharacterByID: Failed to get character by ID %s: %v", idStr, err)
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"data":  char,
//		"count": "1",
//	})
//}
//
//func SaveCharacter(c *gin.Context) {
//	var char character.CharacterConfig
//
//	if err := c.ShouldBindJSON(&char); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
//		log.Printf("SaveCharacter: ShouldBindJSON failed: %v", err)
//		return
//	}
//
//	//Validate Character(char)
//
//	// Generate UUID for the character
//	char.ID = core.NewUUIDv7()
//
//	// Insert to the Database
//	err := character.InsertCustomCharacterConfig(c.Request.Context(), char)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save character: " + err.Error()})
//		log.Printf("SaveCharacter: InsertCustomCharacterConfig failed: %v", err)
//		return
//	}
//
//	c.JSON(http.StatusCreated, gin.H{
//		"data": gin.H{
//			"id": char.ID,
//		},
//		"count": "1",
//	})
//}
