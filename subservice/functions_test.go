package subservice

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"
	"testing"
)

func TestTrimProductLink(t *testing.T) {
	testLinks := []string{
		"https://www.avito.ru/moskva/audio_i_video/ekaterina_lyubimova_-_18_kursov_4_knigi_v_podarok_1906537464",
		"https://www.avito.ru/moskva/audio_i_video/topstretching_rastyazhka_-_bolshoy_nabor_kursov_1970317511",
		"https://www.avito.ru/serpuhov/igry_pristavki_i_programmy/xbox_one_s_1tb_digital_1987431281",
		"https://www.avito.ru/yaroslavl/sport_i_otdyh/eroticheskaya_kartochnaya_igra_1858744046",
		"https://www.avito.ru/ufa/predlozheniya_uslug/banya_i_sauna_mup_lok_zdorove_865588775",
		"https://www.avito.ru/ufa/predlozheniya_uslug/sauna_basseyn_venik_aktsi_ya_1091583802",
		"https://www.avito.ru/serpuhov/igry_pristavki_i_programmy/igrovaya_pristavka_xbox_one_x_1982577979",
		"https://www.avito.ru/serpuhov/igry_pristavki_i_programmy/xbox_one_x_2011215189",
	}
	testResults := []ProductID{
		"1906537464",
		"1970317511",
		"1987431281",
		"1858744046",
		"865588775",
		"1091583802",
		"1982577979",
		"2011215189",
	}

	for idx,testLink := range testLinks {
		if TrimProductLink(testLink) != testResults[idx] {
			t.Errorf("Error! Exptected: %s, got: %s", testResults[idx], TrimProductLink(testLink))
		}
	}
}

func TestSubService_LoadSubMapFromDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		fmt.Println("failed to open sqlmock database:", err)
	}
	defer db.Close()

	ss := NewSubService(db)

	sqlStatement := `
		SELECT \* FROM public."products"
	`
	expectedProductID, expectedSubUsers := "1123123", "{vasya@valera.com, jeylow@yandex.ru}"
	rows := sqlmock.NewRows([]string{"One", "Two"}).AddRow(expectedProductID, expectedSubUsers)
	mock.ExpectQuery(sqlStatement).WillReturnRows(rows)
	ss.LoadSubMapFromDB()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
	if ss.ProductPrices[ProductID(expectedProductID)] != "" {
		t.Error(errors.New("Product Prices Map was not filled!"))
	}
}

func TestSubService_LoadSubMapToDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		fmt.Println("failed to open sqlmock database:", err)
	}
	defer db.Close()

	ss := NewSubService(db)
	sqlStatement := `
		INSERT INTO public."products"
		`
	ss.ProductSubs["123123123"] = []string{"email@mail.com", "email2@mail.com"}
	ss.ProductSubs["9123191"] = []string{"email3@mail.com"}
	mock.ExpectExec(sqlStatement).WithArgs(
		driver.Value("123123123"),
		driver.Value(pq.Array([]string{"email@mail.com", "email2@mail.com"}))).WillReturnResult(driver.ResultNoRows)
	mock.ExpectExec(sqlStatement).WithArgs(
		driver.Value("9123191"),
		driver.Value(pq.Array([]string{"email3@mail.com"}))).WillReturnResult(driver.ResultNoRows)
	ss.LoadSubMapToDB()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

//func TestSubService_AddSubscriberToProduct(t *testing.T) {
//	expectedMap := map[ProductID][]string {
//		"3102301230": []string{"email1@mail.com", "email2@mail.com", "email4@mail.com"},
//		"3231234412": []string{"email1@mail.com", "email2@mail.com", "email5@mail.com"},
//		"12345115": []string{"email1@mail.com"},
//		"5515151515": []string{"email6@mail.com", "email10@mail.com"},
//		"415151612w34": []string{"email23@mail.com", "email15@mail.com", "email5@mail.com"},
//	}
//
//	ss := NewSubService(nil)
//	for key, value := range expectedMap {
//		for _, email:= range value {
//			ss.AddSubscriberToProduct(key, email)
//		}
//	}
//
//	if !reflect.DeepEqual(expectedMap, ss.ProductSubs) {
//		t.Error(errors.New("Maps are not equal!"))
//	}
//
//}

