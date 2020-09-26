package params

// PostSignUp represents request body for POST /api/signup
type PostSignUp struct {
	Email         string `form:"email" validate:"required,email"`
	Password      string `form:"password" validate:"required,min=8"`
	TokenResponse string `form:"token_response" validate:"optional"`
}

func (p *PostSignUp) TrimSpaces() {
}
