package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Joseph-Koop/json-project/internal/data"
	"github.com/Joseph-Koop/json-project/internal/validator"
)

func (a *applicationDependencies) postMemberHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		User_id            int `json:"user_id"`
		Name             string `json:"name"`
		Address          string `json:"address"`
		Phone            int `json:"phone"`
		Email            string `json:"email"`
		Membership_tier            string `json:"membership_tier"`
		Expiry_date            time.Time `json:"expiry_date"`
	}
	
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	member := &data.Member{
		User_id: 		incomingData.User_id,
		Name: 		incomingData.Name,
		Address: 	incomingData.Address,
		Phone: 		incomingData.Phone,
		Email: 		incomingData.Email,
		Membership_tier: 		incomingData.Membership_tier,
		Expiry_date: 		incomingData.Expiry_date,
	}
	
	v := validator.New()
	
	data.ValidateMember(v, member)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) 
		return
	}
	
	err = a.memberModel.Insert(member)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/members/%d", member.ID))

	data := envelope{
		"member": member,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) displayMemberHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	member, err := a.memberModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"member": member,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) updateMemberHandler(w http.ResponseWriter, r *http.Request) {
	
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	member, err := a.memberModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}
	
	var incomingData struct {
		User_id      	*int `json:"user_id"`
		Name      	*string `json:"name"`
		Address     *string `json:"address"`
		Phone      	*int `json:"phone"`
		Email      	*string `json:"email"`
		Membership_tier      	*string `json:"membership_tier"`
		Expiry_date      	*time.Time `json:"expiry_date"`
	}

	
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	if incomingData.User_id != nil {
		member.User_id = *incomingData.User_id
	}
	
	if incomingData.Name != nil {
		member.Name = *incomingData.Name
	}
	
	if incomingData.Address != nil {
		member.Address = *incomingData.Address
	}
	
	if incomingData.Phone != nil {
		member.Phone = *incomingData.Phone
	}
	
	if incomingData.Email != nil {
		member.Email = *incomingData.Email
	}
	
	if incomingData.Membership_tier != nil {
		member.Membership_tier = *incomingData.Membership_tier
	}
	
	if incomingData.Expiry_date != nil {
		member.Expiry_date = *incomingData.Expiry_date
	}
	
	v := validator.New()
	data.ValidateMember(v, member)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	
	err = a.memberModel.Update(member)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"member": member,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) deleteMemberHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.memberModel.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	
	data := envelope{
		"message": "Member successfully deleted.",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listMembersHandler(w http.ResponseWriter, r *http.Request) {
	
	var queryParametersData struct {
		User_id     	*int
		Name     	*string
		Address   	*string
		Phone     	*int
		Email     	*string
		Membership_tier     	*string
		Expiry_date     	*time.Time
		data.Filters
	}
	
	queryParameters := r.URL.Query()

	v := validator.New()

	user_id_string := a.getSingleQueryParameter(queryParameters, "user_id", "")
	if user_id_string != "" {
		user_id_int, err := strconv.Atoi(user_id_string)

		if err == nil && user_id_int != 0 {
			queryParametersData.Phone = &user_id_int
		} else {
			v.AddError(user_id_string, "Must be an integer value.")
		}
	}

	name_string := a.getSingleQueryParameter(queryParameters, "name", "")
	if name_string != "" {
		queryParametersData.Name = &name_string
	}

	address_string := a.getSingleQueryParameter(queryParameters, "address", "")
	if address_string != "" {
		queryParametersData.Address = &address_string
	}

	phone_string := a.getSingleQueryParameter(queryParameters, "phone", "")
	if phone_string != "" {
		phone_int, err := strconv.Atoi(phone_string)

		if err == nil && phone_int != 0 {
			queryParametersData.Phone = &phone_int
		} else {
			v.AddError(phone_string, "Must be an integer value.")
		}
	}

	email_string := a.getSingleQueryParameter(queryParameters, "email", "")
	if email_string != "" {
		queryParametersData.Email = &email_string
	}

	membership_tier_string := a.getSingleQueryParameter(queryParameters, "membership_tier", "")
	if membership_tier_string != "" {
		queryParametersData.Membership_tier = &membership_tier_string
	}

	expiry_date_string := a.getSingleQueryParameter(queryParameters, "expiry_date", "")
	if expiry_date_string != "" {
		parsed_date, err := time.Parse("2006-01-02", expiry_date_string)
		if err != nil {
			v.AddError(expiry_date_string, "Must be an valid date.")
		} else {
			queryParametersData.Expiry_date = &parsed_date
		}
	}

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "name", "address", "phone", "email", "membership_tier", "expiry_date", "-id", "-name", "-address", "-phone", "-email", "-membership_tier", "-expiry_date"}

	
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := a.contextGetUser(r)

	var members []*data.Member
	var err error
	var metadata any

	switch user.Role_id{
		case 1:
			members, metadata, err = a.memberModel.GetAll(queryParametersData.User_id, queryParametersData.Name, queryParametersData.Address, queryParametersData.Phone, queryParametersData.Email, queryParametersData.Membership_tier, queryParametersData.Expiry_date, queryParametersData.Filters)


		case 2:
			members, metadata, err = a.memberModel.GetAll(queryParametersData.User_id, queryParametersData.Name, queryParametersData.Address, queryParametersData.Phone, queryParametersData.Email, queryParametersData.Membership_tier, queryParametersData.Expiry_date, queryParametersData.Filters)

		case 3:
			member, err2 := a.memberModel.GetByUserID(user.ID)
			if err2 != nil {
				a.serverErrorResponse(w, r, err2)
				return
			}

			members, metadata, err = a.memberModel.GetAllByMemberID(member.ID, queryParametersData.User_id, queryParametersData.Name, queryParametersData.Address, queryParametersData.Phone, queryParametersData.Email, queryParametersData.Membership_tier, queryParametersData.Expiry_date, queryParametersData.Filters)
		default:
			a.notPermittedResponse(w, r)
			return
	}
	
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"members":   members,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
