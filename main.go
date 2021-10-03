package main

import (
	"database/sql"  // Funções de manipulação de bancos de dados.
	"log"           // Funções de impressão de mensagens no console.
	"net/http"      // Funções de requisição, URLs e servidor web.
	"regexp"        // Funções para busca de padrões em strings.
	"text/template" // Funções de gerenciamento de templates.

	_ "github.com/go-sql-driver/mysql" // Driver MySQL para Go.
)

// BD em uso.
var currentDB string = "default"

// Estrutura dos BDs.
type Database struct {
	Name string
}

var IsAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// Conexão com o BD.
func dbConn() (db *sql.DB) {
	dbDriver := "mysql" // SGBD.
	dbUser := "root"    // Usuário.
	dbPass := ""        // Senha.
	dbName := currentDB // BD.

	// Abre uma conexão com o BD.
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	// Verifica se houve erros.
	if err != nil {
		panic(err.Error())
	}
	return db
}

// Rederização dos templates da pasta tmpl.
var tmpl = template.Must(template.ParseGlob("tmpl/*"))

// Pegando todos os BDs cadastrados e indo para a página Index.
func Index(w http.ResponseWriter, r *http.Request) {
	log.Println("Abrindo página index...")
	log.Println(currentDB)
	// Abre a conexão.
	db := dbConn()
	// Consulta.
	selDB, err := db.Query("SHOW DATABASES")
	// Verifica se houve erros.
	if err != nil {
		panic(err.Error())
	}
	// Cria BD.
	database := Database{}
	// Array de BDs.
	res := []Database{}
	// Percorre cada linha retornada na consulta.
	for selDB.Next() {
		// Variável que vai receber o nome de cada BD.
		var name string
		// Faz o Scan do SELECT.
		err = selDB.Scan(&name)
		// Verifica se houve erros.
		if err != nil {
			panic(err.Error())
		}
		// Cria um objeto Database.
		database.Name = name
		// Adiciona o objeto no array.
		res = append(res, database)
	}
	// Vai para a página Index.
	tmpl.ExecuteTemplate(w, "Index", res)
	// Fecha a conexão.
	db.Close()
}

// Criando um novo BD.
func CreateDB(w http.ResponseWriter, r *http.Request) {
	log.Println("Criando um banco de dados...")
	// Abre a conexão.
	db := dbConn()
	// Verifica o método de requisição.
	if r.Method == "POST" {
		// Pega o valor do campo "dbNome" do formulário.
		dbName := r.FormValue("db-name")
		if IsAlphaNumeric(dbName) {
			// Executa o comando SQL.
			_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
			// Verificar se houver erros.
			if err != nil {
				panic(err.Error())
			}
			// Volta para a página Index.
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		} else {
			// Volta para a página Index informando que o nome passado é inválido.
			http.Redirect(w, r, "/?err=invalide-db-name", http.StatusMovedPermanently)
		}
	}
}

// Mudando de BD.
func UseDB(w http.ResponseWriter, r *http.Request) {
	log.Println("Mudando de BD...")
	// Mudando o valor da variável controladora.
	currentDB = r.URL.Query().Get("db")
	log.Println(currentDB)
	// Redirecionando para o Index.
	http.Redirect(w, r, "/?db="+currentDB, http.StatusMovedPermanently)
}

// Removendo um BD.
func DropDB(w http.ResponseWriter, r *http.Request) {
	// Abre a conexão.
	db := dbConn()
	// Pegando o nome do banco pela URL.
	dbName := r.URL.Query().Get("db-name")
	// Executa comando de remoção.
	_, err := db.Exec("DROP DATABASE " + dbName)
	// Verifica se houve erros.
	if err != nil {
		panic(err.Error())
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func main() {
	// Informa que o servidor está no ar.
	log.Println("Server started on: http://localhost:8080")

	// Gerenciamento das URLs.
	http.HandleFunc("/", Index)
	http.HandleFunc("/create-db", CreateDB)
	http.HandleFunc("/use-db", UseDB)
	http.HandleFunc("/drop-db", DropDB)

	// Inicia o servidor.
	http.ListenAndServe("localhost:9002", nil)
}
