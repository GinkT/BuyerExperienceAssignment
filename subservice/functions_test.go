package subservice

import "testing"

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
	testResults := []productID{
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

