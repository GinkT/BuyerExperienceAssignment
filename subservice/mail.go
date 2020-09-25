package subservice

import (
	"log"
	"net/smtp"
)

// Отправляет email с обновленной стоимостью всем подписчикам из emails[]
func (SubServ *SubService)SendMailToFollowers(id ProductID, cost string, emails []string) {
	message := []byte("From: \"Сервис подписки\" <avitobuyerexperience@yandex.ru>\r\n" +
		"Subject: Изменение цены\r\n" +
		"\r\n" +
		"Здравствуйте, цена товара, на который вы подписались (https://www.avito.ru/"+ string(id) + ") изменилась и составляет: " + cost)

	if err := smtp.SendMail("smtp.yandex.ru:587", SubServ.mailerAuth, "avitobuyerexperience@yandex.ru", emails, message); err != nil {
		log.Println("[mail] Error SendMail: ", err)
		return
	}
	log.Println("[mail] Send sub-change emails to:", emails)
}

// Отправляет письмо. Внутри письма ссылка с подтверждением
func (SubServ *SubService)SendConfirmationEmail(link, email string) {
	message := []byte("From: \"Сервис подписки\" <avitobuyerexperience@yandex.ru>\r\n" +
		"Subject: Подтверждение подписки\r\n" +
		"\r\n" +
		"Здравствуйте, подтвердите вашу подписку на товар перейдя по ссылке, указанной в письме:\r\n" +
		"http://localhost:8181/subscribe?link=" + link + "&mail=" + email + "&code=" + SubServ.ConfirmCode + "\r\n")

	if err := smtp.SendMail("smtp.yandex.ru:587", SubServ.mailerAuth, "avitobuyerexperience@yandex.ru", []string{email}, message); err != nil {
		log.Println("[mail] Error SendMail: ", err)
		return
	}
	log.Println("[mail] Send confirm email to:", email)
}
