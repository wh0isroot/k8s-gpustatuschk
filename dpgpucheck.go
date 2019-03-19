package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/smtp"
	"os"
	"reflect"
	"strconv"
)

var data map[string]interface{}

type loginAuth struct {
	user string
	pwd  string
}

func newloginAuth(user, pwd string) smtp.Auth {
	return loginAuth{user: user, pwd: pwd}
}

func (a loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.user), nil
		case "Password:":
			return []byte(a.pwd), nil
		default:
			return nil, errors.New("Unknown fromServer")
		}
	}
	return nil, nil
}

func sendMail(gpunum int) {
	nodename := os.Getenv("NODENAME")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 587
	smtpLogin := os.Getenv("SMTP_LOGIN")
	smtpPasswd := os.Getenv("SMTP_PASSWD")

	useTLS := false
	useStartTLS := true

	from := os.Getenv("SMTP_FROM")
	to := os.Getenv("SMTP_TO")
	title := "GPU NUM ERROR"

	body := "HOSTNAME: " + nodename + "\n" +
		"GPU NUM: " + strconv.Itoa(gpunum)

	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = title
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	conn, err := net.Dial("tcp", smtpHost+":"+strconv.Itoa(smtpPort))
	if err != nil {
		log.Panic(err)
		return
	}

	// TLS
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpHost,
	}

	if useTLS {
		conn = tls.Client(conn, tlsconfig)
	}

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		log.Panic(err)
		return
	}

	hasStartTLS, _ := client.Extension("STARTTLS")
	if useStartTLS && hasStartTLS {
		if err = client.StartTLS(tlsconfig); err != nil {
			log.Panic(err)
			return
		}
	}

	// Set up authentication information.
	auth := smtp.Auth(newloginAuth(smtpLogin, smtpPasswd))

	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			fmt.Printf("Error during AUTH %s\n", err)
			return
		}
	}

	if err := client.Mail(from); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	if err := client.Rcpt(to); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	w, err := client.Data()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	err = w.Close()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	client.Quit()

}

func main() {
	b, err := ioutil.ReadFile("/var/lib/kubelet/device-plugins/kubelet_internal_checkpoint")
	if err != nil {
		panic(err)
	}
	if json.Unmarshal([]byte(b), &data) != nil {
		panic(err)
	}
	registeredDevices, ok := data["RegisteredDevices"].(map[string]interface{})
	if !ok {
		panic("RegisteredDevices is not a map!")
	}
	gpus := reflect.ValueOf(registeredDevices["nvidia.com/gpu"])
	num := gpus.Len()
	if num < 8 {
		sendMail(num)
		panic(num)
	}
}
