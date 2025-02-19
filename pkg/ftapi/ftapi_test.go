package ftapi

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/clientcredentials"
)

func TestNew(t *testing.T) {
	ftAPI := New("https://api.intra.42.fr", &http.Client{})
	assert.IsType(t, ftAPI, &API{}, ftAPI)
}

func TestNewFromCredentials(t *testing.T) {
	ftAPI := NewFromCredentials("https://api.intra.42.fr", &clientcredentials.Config{})
	assert.IsType(t, ftAPI, &API{}, ftAPI)
}

func getBody(body io.ReadCloser) string {
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

func TestGet(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "/v2/users", req.URL.String())
		rw.Header().Add("X-Test", "test_value")
		_, _ = rw.Write([]byte(`OK`))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	resp, err := ftAPI.Get("/v2/users")
	assert.Nil(t, err)
	assert.Equal(t, "OK", getBody(resp.Body))
	assert.Equal(t, "test_value", resp.Header.Get("X-Test"))
}

func TestPost(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "/v1/users", req.URL.String())
		assert.Equal(t, "custom/content-type", req.Header.Get("Content-Type"))
		assert.Equal(t, "this_is_the_body", getBody(req.Body))
		rw.Header().Add("X-Test", "test_value")
		_, _ = rw.Write([]byte(`OK`))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	body := bytes.NewReader([]byte("this_is_the_body"))
	resp, err := ftAPI.Post("/v1/users", "custom/content-type", body)
	assert.Nil(t, err)
	assert.Equal(t, "OK", getBody(resp.Body))
	assert.Equal(t, "test_value", resp.Header.Get("X-Test"))
}

func TestPatch(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "PATCH", req.Method)
		assert.Equal(t, "/v1/users", req.URL.String())
		assert.Equal(t, "custom/content-type", req.Header.Get("Content-Type"))
		assert.Equal(t, "this_is_the_body", getBody(req.Body))
		rw.Header().Add("X-Test", "test_value")
		_, _ = rw.Write([]byte(`OK`))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	body := bytes.NewReader([]byte("this_is_the_body"))
	resp, err := ftAPI.Patch("/v1/users", "custom/content-type", body)
	assert.Nil(t, err)
	assert.Equal(t, "OK", getBody(resp.Body))
	assert.Equal(t, "test_value", resp.Header.Get("X-Test"))
}

type testData struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Array []string `json:"array"`
}

func TestPostJson(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "/v1/users", req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t, "{\"id\":10,\"name\":\"Spoody\",\"array\":[\"test_1\",\"test_2\"]}", getBody(req.Body))
		rw.Header().Add("X-Test", "test_value")
		_, _ = rw.Write([]byte(`OK`))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	resp, err := ftAPI.PostJSON("/v1/users", testData{
		ID:    10,
		Name:  "Spoody",
		Array: []string{"test_1", "test_2"},
	})
	assert.Nil(t, err)
	assert.Equal(t, "OK", getBody(resp.Body))
	assert.Equal(t, "test_value", resp.Header.Get("X-Test"))
}

func TestPatchJson(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "PATCH", req.Method)
		assert.Equal(t, "/v1/users", req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t, "{\"id\":10,\"name\":\"Spoody\",\"array\":[\"test_1\",\"test_2\"]}", getBody(req.Body))
		rw.Header().Add("X-Test", "test_value")
		_, _ = rw.Write([]byte(`OK`))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	resp, err := ftAPI.PatchJSON("/v1/users", testData{
		ID:    10,
		Name:  "Spoody",
		Array: []string{"test_1", "test_2"},
	})
	assert.Nil(t, err)
	assert.Equal(t, "OK", getBody(resp.Body))
	assert.Equal(t, "test_value", resp.Header.Get("X-Test"))
}

func TestHourlyLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "/v2/users", req.URL.String())
		rw.Header().Add("X-Hourly-Ratelimit-Remaining", "0")
		rw.WriteHeader(http.StatusTooManyRequests)
		_, _ = rw.Write([]byte(`Not ok`))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	resp, err := ftAPI.Get("/v2/users")
	assert.NotNil(t, err)
	assert.Equal(t, "exceeded rate limit", err.Error())
	assert.Nil(t, resp)
}

func TestCreateUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "/users", req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t,
			"{\"user\":{\"campus_id\":21,\"email\":\"spoody@test.local\",\"first_name\":\"Spooder\",\"kind\":\"admin\",\"last_name\":\"Webz\",\"login\":\"spoody\"}}",
			getBody(req.Body),
		)
		rw.WriteHeader(http.StatusCreated)
		_, _ = rw.Write([]byte("{\"id\": 127,\"login\":\"spoody\",\"url\": \"https://api.intra.42.fr/v2/users/spoody\"}"))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	user := User{
		Login:     "spoody",
		Email:     "spoody@test.local",
		FirstName: "Spooder",
		LastName:  "Webz",
		Kind:      "admin",
	}
	expectedUser := user
	expectedUser.ID = 127
	expectedUser.URL = "https://api.intra.42.fr/v2/users/spoody"
	err := ftAPI.CreateUser(&user, 21)
	assert.Nil(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestSetUserImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseMultipartForm(10 << 20)
		assert.Nil(t, err)
		assert.Equal(t, "PATCH", req.Method)
		assert.Equal(t, "/users/spoody", req.URL.String())
		assert.Contains(t, req.Header.Get("Content-Type"), "multipart/form-data; boundary=")
		file, fileHeader, err := req.FormFile("user[image]")
		assert.Nil(t, err)
		assert.NotNil(t, file)
		assert.NotNil(t, fileHeader)
		assert.Equal(t, "profile_photo.png", fileHeader.Filename)
		assert.Equal(t, int64(99412), fileHeader.Size)
		assert.Equal(t, "form-data; name=\"user[image]\"; filename=\"profile_photo.png\"", fileHeader.Header.Get("Content-Disposition"))
		rw.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	imgFile, err := os.Open("../../tests/profile_photo.png")
	if err != nil {
		t.Fatal(err)
	}
	ftAPI := New(server.URL, server.Client())
	err = ftAPI.SetUserImage("spoody", imgFile)
	imgFile.Close()
	assert.Nil(t, err)
}

func TestSetUserImageWithNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	imgFile, err := os.Open("../../tests/profile_photo.png")
	if err != nil {
		t.Fatal(err)
	}
	defer imgFile.Close()
	ftAPI := New(server.URL, server.Client())
	err = ftAPI.SetUserImage("spoody", imgFile)
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestSetUserImageWithFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	imgFile, err := os.Open("../../tests/profile_photo.png")
	if err != nil {
		t.Fatal(err)
	}
	defer imgFile.Close()
	ftAPI := New(server.URL, server.Client())
	err = ftAPI.SetUserImage("spoody", imgFile)
	assert.NotNil(t, err)
	assert.Equal(t, "failed setting profile image", err.Error())
}

func TestCreateClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "/users/spoody/closes", req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t,
			"{\"close\":{\"closer_id\":37,\"kind\":\"agu\",\"reason\":\"This is for testing purposes\"}}",
			getBody(req.Body),
		)
		rw.WriteHeader(http.StatusCreated)
		_, _ = rw.Write([]byte("{\"id\":13,\"reason\":\"This is for testing purposes\",\"state\":\"close\",\"created_at\":\"2017-11-22T13:43:29.676Z\",\"updated_at\":\"2017-11-22T13:43:29.676Z\",\"community_services\":[],\"user\":{\"id\":37,\"login\":\"ebou-nya\",\"url\":\"https://api.intra.42.fr/v2/users/ebou-nya\"},\"closer\":{\"id\":42,\"login\":\"spoody\",\"url\":\"https://api.intra.42.fr/v2/users/spoody\"}}"))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	err := ftAPI.CreateClose(&Close{
		Kind:   "agu",
		Reason: "This is for testing purposes",
		User: &User{
			Login: "spoody",
		},
		Closer: &User{
			ID: 37,
		},
	})
	assert.Nil(t, err)
}

func TestGetUserByLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "/users/spoody", req.URL.String())
		assert.Equal(t, "", getBody(req.Body))
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("{\"id\":66356,\"email\":\"mehdi@1337.ma\",\"login\":\"spoody\",\"first_name\":\"Mehdi\",\"last_name\":\"Bounya\",\"usual_first_name\":null,\"url\":\"https://api.intra.42.fr/v2/users/spoody\",\"phone\":\"hidden\",\"displayname\":\"Mehdi Bounya\",\"usual_full_name\":\"Mehdi Bounya\",\"image_url\":\"https://cdn.intra.42.fr/users/spoody.png\",\"staff?\":true,\"correction_point\":5,\"pool_month\":\"april\",\"pool_year\":\"2019\",\"location\":null,\"wallet\":0,\"anonymize_date\":\"2022-01-22T00:00:00.000+01:00\",\"groups\":[],\"cursus_users\":[{\"grade\":null,\"level\":0,\"skills\":[],\"blackholed_at\":null,\"id\":95640,\"begin_at\":\"2020-06-20T20:33:00.000Z\",\"end_at\":null,\"cursus_id\":1,\"has_coalition\":true,\"user\":{\"id\":66356,\"login\":\"spoody\",\"url\":\"https://api.intra.42.fr/v2/users/spoody\"},\"cursus\":{\"id\":1,\"created_at\":\"2014-11-02T16:43:38.480Z\",\"name\":\"42\",\"slug\":\"42\"}},{\"grade\":null,\"level\":0,\"skills\":[],\"blackholed_at\":null,\"id\":95639,\"begin_at\":\"2020-06-20T20:27:00.000Z\",\"end_at\":null,\"cursus_id\":4,\"has_coalition\":true,\"user\":{\"id\":66356,\"login\":\"spoody\",\"url\":\"https://api.intra.42.fr/v2/users/spoody\"},\"cursus\":{\"id\":4,\"created_at\":\"2015-05-01T17:46:08.433Z\",\"name\":\"Piscine C\",\"slug\":\"piscine-c\"}},{\"grade\":null,\"level\":0,\"skills\":[],\"blackholed_at\":\"2020-01-15T23:00:00.000Z\",\"id\":79645,\"begin_at\":\"2019-10-31T13:15:00.000Z\",\"end_at\":null,\"cursus_id\":21,\"has_coalition\":true,\"user\":{\"id\":66356,\"login\":\"spoody\",\"url\":\"https://api.intra.42.fr/v2/users/spoody\"},\"cursus\":{\"id\":21,\"created_at\":\"2019-07-29T08:45:17.896Z\",\"name\":\"42cursus\",\"slug\":\"42cursus\"}}],\"projects_users\":[],\"languages_users\":[{\"id\":320755,\"language_id\":2,\"user_id\":66356,\"position\":1,\"created_at\":\"2020-11-30T15:14:52.040Z\"},{\"id\":320756,\"language_id\":1,\"user_id\":66356,\"position\":2,\"created_at\":\"2020-11-30T15:14:52.140Z\"}],\"achievements\":[{\"id\":218,\"name\":\"Welcome, Learner!\",\"description\":\"Tu as réussi ta piscine C. Bienvenue à 42 !\",\"tier\":\"none\",\"kind\":\"project\",\"visible\":true,\"image\":\"/uploads/achievement/image/218/PRO001-2.svg\",\"nbr_of_success\":null,\"users_url\":\"https://api.intra.42.fr/v2/achievements/218/users\"}],\"titles\":[{\"id\":126,\"name\":\"The true and only %login\"}],\"titles_users\":[{\"id\":4844,\"user_id\":66356,\"title_id\":126,\"selected\":true}],\"partnerships\":[],\"patroned\":[],\"patroning\":[],\"expertises_users\":[{\"id\":46990,\"expertise_id\":15,\"interested\":false,\"value\":2,\"contact_me\":false,\"created_at\":\"2020-08-19T20:12:53.496Z\",\"user_id\":66356}],\"roles\":[{\"id\":2,\"name\":\"Events Manager\"}],\"campus\":[{\"id\":21,\"name\":\"Benguerir\",\"time_zone\":\"Africa/Casablanca\",\"language\":{\"id\":1,\"name\":\"Français\",\"identifier\":\"fr\",\"created_at\":\"2014-11-02T16:43:38.466Z\",\"updated_at\":\"2021-01-19T04:13:46.180Z\"},\"users_count\":982,\"vogsphere_id\":11,\"country\":\"Morocco\",\"address\":\"Lot 660, Hay Moulay Rachid, Ben Guerir 43150\",\"zip\":\"43150\",\"city\":\"Benguerir\",\"website\":\"https://1337.ma\",\"facebook\":\"\",\"twitter\":\"\",\"active\":true,\"email_extension\":\"1337.ma\"}],\"campus_users\":[{\"id\":57173,\"user_id\":66356,\"campus_id\":21,\"is_primary\":true}]}"))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	user, err := ftAPI.GetUserByLogin("spoody")
	assert.NotNil(t, user)
	assert.Nil(t, err)
	assert.Equal(t, 66356, user.ID)
	assert.Equal(t, "spoody", user.Login)
	assert.Equal(t, "mehdi@1337.ma", user.Email)
	assert.Equal(t, "Mehdi", user.FirstName)
	assert.Equal(t, "Bounya", user.LastName)
	assert.Equal(t, "", user.UsualFirstName)
	assert.Equal(t, "hidden", user.Phone)
	assert.True(t, user.IsStaff)
	assert.Equal(t, "https://api.intra.42.fr/v2/users/spoody", user.URL)
	assert.Equal(t, "april", user.PoolMonth)
	assert.Equal(t, "2019", user.PoolYear)
	assert.NotNil(t, user.Campuses)
	assert.NotNil(t, user.CampusUsers)
	assert.Len(t, user.Campuses, 1)
	assert.Equal(t, 21, user.Campuses[0].ID)
	assert.Equal(t, "Benguerir", user.Campuses[0].Name)
	assert.Equal(t, "Africa/Casablanca", user.Campuses[0].TimeZone)
	assert.Equal(t, 982, user.Campuses[0].UsersCount)
	assert.Equal(t, 11, user.Campuses[0].VogsphereID)

	assert.Equal(t, 1, user.Campuses[0].Language.ID)
	assert.Equal(t, "Français", user.Campuses[0].Language.Name)
	assert.Equal(t, "fr", user.Campuses[0].Language.ISOIdentifier)
	assert.Equal(t, "2014-11-02 16:43:38.466 +0000 UTC", user.Campuses[0].Language.CreatedAt.String())
	assert.Equal(t, "2021-01-19 04:13:46.18 +0000 UTC", user.Campuses[0].Language.UpdatedAt.String())

	assert.Len(t, user.CampusUsers, 1)
	assert.Equal(t, 57173, user.CampusUsers[0].ID)
	assert.Equal(t, 66356, user.CampusUsers[0].UserID)
	assert.Equal(t, 21, user.CampusUsers[0].CampusID)
	assert.True(t, user.CampusUsers[0].IsPrimary)

	primaryCampus := user.GetPrimaryCampus()
	assert.NotNil(t, primaryCampus)
	assert.Equal(t, user.Campuses[0], primaryCampus)

}

func TestUpdateUser(t *testing.T) {
	testData := []map[string]interface{}{
		{
			"payload": &User{
				Email:     "spoody@local.test",
				FirstName: "Spooder",
				LastName:  "Webz",
				Password:  "test",
				Kind:      "student",
			},
			"expected_payload": "{\"user\":{\"email\":\"spoody@local.test\",\"first_name\":\"Spooder\",\"kind\":\"student\",\"last_name\":\"Webz\",\"password\":\"test\"}}",
		},
		{
			"payload": &User{
				Email: "spoody@local.test",
			},
			"expected_payload": "{\"user\":{\"email\":\"spoody@local.test\"}}",
		},
		{
			"payload": &User{
				FirstName: "Spooder",
			},
			"expected_payload": "{\"user\":{\"first_name\":\"Spooder\"}}",
		},
		{
			"payload": &User{
				LastName: "Webz",
			},
			"expected_payload": "{\"user\":{\"last_name\":\"Webz\"}}",
		},
		{
			"payload": &User{
				Password: "test",
			},
			"expected_payload": "{\"user\":{\"password\":\"test\"}}",
		},
		{
			"payload": &User{
				Kind: "student",
			},
			"expected_payload": "{\"user\":{\"kind\":\"student\"}}",
		},
	}
	for _, val := range testData {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "PATCH", req.Method)
			assert.Equal(t, "/users/spoody", req.URL.String())
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			assert.Equal(t,
				val["expected_payload"],
				getBody(req.Body),
			)
			rw.WriteHeader(http.StatusNoContent)
			_, _ = rw.Write([]byte(""))
		}))
		ftAPI := New(server.URL, server.Client())
		err := ftAPI.UpdateUser("spoody", val["payload"].(*User))
		assert.Nil(t, err)
		server.Close()
	}

}

func TestAddCorrectionPoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "/users/spoody/correction_points/add", req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t,
			"{\"amount\":5,\"reason\":\"Testing\"}",
			getBody(req.Body),
		)
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(""))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	err := ftAPI.AddCorrectionPoints("spoody", 5, "Testing")
	assert.Nil(t, err)
}

func TestRemoveCorrectionPoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "DELETE", req.Method)
		assert.Equal(t, "/users/spoody/correction_points/remove", req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t,
			"{\"amount\":5,\"reason\":\"Testing\"}",
			getBody(req.Body),
		)
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(""))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	err := ftAPI.RemoveCorrectionPoints("spoody", 5, "Testing")
	assert.Nil(t, err)
}

func TestGetProjectByName(t *testing.T) {
	assert := assert.New(t)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal("GET", req.Method)
		assert.Equal("/projects/libft", req.URL.String())
		assert.Equal("", getBody(req.Body))
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("{\"id\": 1,\"name\": \"Libft\",\"slug\": \"libft\",\"parent\": null,\"children\": [],\"attachments\": [],\"created_at\": \"2014-11-02T18:23:57.156Z\",\"updated_at\": \"2021-11-03T09:36:15.705Z\",\"exam\": false,\"git_id\": null,\"repository\": null,\"cursus\": [{\"id\": 1,\"created_at\": \"2014-11-02T16:43:38.480Z\",\"name\": \"42\",\"slug\": \"42\"},{\"id\": 8,\"created_at\": \"2015-09-22T11:35:20.216Z\",\"name\": \"WeThinkCode_\",\"slug\": \"wethinkcode_\"},{\"id\": 10,\"created_at\": \"2015-11-12T16:35:19.549Z\",\"name\": \"Formation Pole Emploi\",\"slug\": \"formation-pole-emploi\"}],\"campus\": [{\"id\": 42,\"name\": \"42Network\",\"time_zone\": \"Europe/Paris\",\"language\": {\"id\": 2,\"name\": \"English\",\"identifier\": \"en\",\"created_at\": \"2015-04-14T16:07:38.122Z\",\"updated_at\": \"2021-11-05T14:28:42.585Z\"},\"users_count\": 12,\"vogsphere_id\": 1,\"country\": \"France\",\"address\": \"96 boulevard Bessières\",\"zip\": \"75017\",\"city\": \"PARIS\",\"website\": \"http://42network.org/\",\"facebook\": \"\",\"twitter\": \"\",\"active\": true,\"email_extension\": \"42network.org\",\"default_hidden_phone\": false},{\"id\": 26,\"name\": \"Tokyo\",\"time_zone\": \"Asia/Tokyo\",\"language\": {\"id\": 13,\"name\": \"Japanese\",\"identifier\": \"ja\",\"created_at\": \"2019-11-15T13:34:10.581Z\",\"updated_at\": \"2021-11-03T12:42:58.540Z\"},\"users_count\": 3184,\"vogsphere_id\": 17,\"country\": \"Japan\",\"address\": \"Sumitomo Fudosan Roppongi Grand Tower 3-2-1 Roppongi Minato-ku reception: 24F\",\"zip\": \"106-6224\",\"city\": \"Tokyo\",\"website\": \"https://42tokyo.jp\",\"facebook\": \"https://www.facebook.com/42tokyo/\",\"twitter\": \"https://twitter.com/42_tokyo\",\"active\": true,\"email_extension\": \"42tokyo.jp\",\"default_hidden_phone\": true}],\"videos\": [],\"project_sessions\": [{\"id\": 2697,\"solo\": true,\"begin_at\": null,\"end_at\": null,\"estimate_time\": \"14 days\",\"difficulty\": 85,\"objectives\": [\"Basics of C programming\",\"Unix C library\",\"Static library creation\"],\"description\": \"This project is going to help you consolidate your piscine experience. You are going to re-code several functions of the standard C library, as well as other utility functions that you will use often during your training.\",\"duration_days\": null,\"terminating_after\": null,\"project_id\": 1,\"campus_id\": 13,\"cursus_id\": 1,\"created_at\": \"2018-09-06T13:41:29.663Z\",\"updated_at\": \"2021-09-24T08:18:18.319Z\",\"max_people\": null,\"is_subscriptable\": true,\"scales\": [{\"id\": 826,\"correction_number\": 5,\"is_primary\": true},{\"id\": 697,\"correction_number\": 5,\"is_primary\": false},{\"id\": 672,\"correction_number\": 5,\"is_primary\": false},{\"id\": 1,\"correction_number\": 5,\"is_primary\": false}],\"uploads\": [{\"id\": 70,\"name\": \"Moulinette\"}],\"team_behaviour\": \"user\",\"commit\": null},{\"id\": 944,\"solo\": true,\"begin_at\": null,\"end_at\": null,\"estimate_time\": \"7 days\",\"difficulty\": 100,\"objectives\": [\"Basics of C programming\",\"Unix C library\",\"Static library creation\"],\"description\": \"Ce premier projet en tant qu'étudiant de 42 va vous faire consolider vos acquis de piscine. Vous allez recoder un certain nombre de fonctions de la librairie C standard, ainsi que d'autres fonctions utilitaires que vous réutiliserez tout au long de votre cursus.\",\"duration_days\": null,\"terminating_after\": null,\"project_id\": 1,\"campus_id\": 7,\"cursus_id\": 1,\"created_at\": \"2016-10-31T16:23:40.149Z\",\"updated_at\": \"2020-04-08T03:17:41.804Z\",\"max_people\": null,\"is_subscriptable\": false,\"scales\": [{\"id\": 826,\"correction_number\": 5,\"is_primary\": true},{\"id\": 697,\"correction_number\": 5,\"is_primary\": false},{\"id\": 672,\"correction_number\": 5,\"is_primary\": false},{\"id\": 1,\"correction_number\": 5,\"is_primary\": false}],\"uploads\": [{\"id\": 70,\"name\": \"Moulinette\"}],\"team_behaviour\": \"user\",\"commit\": null}]}"))
	}))
	defer server.Close()
	ftAPI := New(server.URL, server.Client())
	project, err := ftAPI.GetProjectByName("libft")

	assert.NotNil(project)
	assert.Nil(err)
	assert.Equal(1, project.ID)
	assert.Equal("Libft", project.Name)
	assert.Equal("libft", project.Slug)
	assert.Equal("2014-11-02 18:23:57.156 +0000 UTC", project.CreatedAt.String())
	assert.Equal("2021-11-03 09:36:15.705 +0000 UTC", project.UpdatedAt.String())
	assert.False(project.Exam)
	assert.Nil(project.GitID)
	assert.Nil(project.Repogitory)

	assert.Equal(1, project.Cursus[0].ID)
	assert.Equal("2014-11-02 16:43:38.48 +0000 UTC", project.Cursus[0].CreatedAt.String())
	assert.Equal("42", project.Cursus[0].Name)
	assert.Equal("42", project.Cursus[0].Slug)

	assert.Equal(42, project.Campus[0].ID)
	assert.Equal("42Network", project.Campus[0].Name)
	assert.Equal("Europe/Paris", project.Campus[0].TimeZone)
	// TODO
	// more test for Campus

	assert.Equal([]string{}, project.Videos)

	assert.Equal(2697, project.ProjectSessions[0].ID)
	assert.True(project.ProjectSessions[0].Solo)
	assert.Equal("14 days", project.ProjectSessions[0].EstimateTime)
	assert.Equal(85, project.ProjectSessions[0].Difficulty)
	assert.Equal("Basics of C programming", project.ProjectSessions[0].Objectives[0])
	assert.Equal("Unix C library", project.ProjectSessions[0].Objectives[1])
	assert.Equal("Static library creation", project.ProjectSessions[0].Objectives[2])
	assert.Equal(1, project.ProjectSessions[0].ProjectID)
	assert.Equal(13, project.ProjectSessions[0].CampusID)
	// TODO
	// more test for ProjectSessions
}
