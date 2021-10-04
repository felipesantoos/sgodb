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

// Estrutura das tabelas.
type Table struct {
	Name string
}

var IsValid = regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString

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
		if IsValid(dbName) {
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
	// Fecha a conexão.
	defer db.Close()
}

// Consulta as tabelas do banco selecionado.
func SelectTables(w http.ResponseWriter, currentDB string) {
	// Abre a conexão.
	db := dbConn()
	// Executa o comando SQL.
	selDB, err := db.Query("SHOW TABLES")
	// Verifica se houve erros.
	if err != nil {
		panic(err.Error())
	}
	// Variável que vai ser passada para a próxima página.
	table := Table{}
	// Array de tabelas.
	res := []Table{}
	// Percorre os dados retornados.
	for selDB.Next() {
		// Variável que vai armazenar os nomes das tabelas.
		var name string
		// Faz o Scan do SHOW TABLES.
		err = selDB.Scan(&name)
		// Verifica se houve erros.
		if err != nil {
			panic(err.Error())
		}
		// Salva o nome da tabela na struc table.
		table.Name = name
		log.Println(table.Name)
		// Adicionar a tabela no array.
		res = append(res, table)
	}
	var dbAndTbs = []interface{}{currentDB, res}
	// Vai para outra página.
	tmpl.ExecuteTemplate(w, "UseDB", dbAndTbs)
	// Fecha a conexão.
	defer db.Close()
}

// Muda de BD.
func UseDB(w http.ResponseWriter, r *http.Request) {
	log.Println("Mudando de BD...")
	// Mudando o valor da variável controladora.
	currentDB = r.URL.Query().Get("db")
	log.Println(currentDB)
	// Redirecionando para a página do banco de dados.
	SelectTables(w, currentDB)
}

// Remove um BD.
func DropDB(w http.ResponseWriter, r *http.Request) {
	// Abre a conexão.
	db := dbConn()
	// Pegando o nome do banco pela URL.
	dbName := r.URL.Query().Get("db-name")
	// Executa comando de remoção.
	_, err := db.Exec("DROP DATABASE `" + dbName + "`")
	// Verifica se houve erros.
	if err != nil {
		panic(err.Error())
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
	defer db.Close()
}

// Cria uma tabela.
func CreateTable(w http.ResponseWriter, r *http.Request) {
	// Abre a conexão.
	db := dbConn()
	// Verifica qual é o método da requisição.
	if r.Method == "POST" {
		// Pega os dados do formulário.
		database := r.FormValue("db")
		table := r.FormValue("table-name")
		// Verifica se o nome é válido.
		if IsValid(table) {
			// Executa o comando SQL.
			_, err := db.Exec("CREATE TABLE `" + table + "` (id INT AUTO_INCREMENT NOT NULL PRIMARY KEY)")
			// Verifica se houve erros.
			if err != nil {
				panic(err.Error())
			}
			// Redireciona para a página do BD.
			http.Redirect(w, r, "use-db?db="+database, http.StatusMovedPermanently)
		} else {
			log.Println("Nome de tabela inválido!")
		}
	}
	// Fecha a conexão.
	defer db.Close()
}

// Remove uma tabela do banco selecionado.
func DropTable(w http.ResponseWriter, r *http.Request) {
	log.Println("Apague")
	// Abre a conexão.
	db := dbConn()
	database := r.URL.Query().Get("db")
	table := r.URL.Query().Get("table")
	// Executa o comando SQL.
	_, err := db.Exec("DROP TABLE `" + table + "`")
	// Verifica se houve erros.
	if err != nil {
		panic(err.Error())
	}
	// Redireciona para a página do BD.
	http.Redirect(w, r, "use-db?db="+database, http.StatusMovedPermanently)
	// Fecha a conexão.
	defer db.Close()
}

func main() {
	// Informa que o servidor está no ar.
	log.Println("Server started on: http://localhost:8080")

	// Gerenciamento das URLs.
	// A URL localhost:8080/ executa a função Index.
	http.HandleFunc("/", Index)
	// A URL localhost:8080/create-db executa a função CreateDB.
	http.HandleFunc("/create-db", CreateDB)
	// A URL localhost:8080/use-db executa a função UseDB.
	http.HandleFunc("/use-db", UseDB)
	// A URL localhost:8080/drop-db executa a função DropDB.
	http.HandleFunc("/drop-db", DropDB)
	// A URL localhost:8080/create-table executa a função CreateTable.
	http.HandleFunc("/create-table", CreateTable)
	// A URL localhost:8080/drop-table executa a função DropTable.
	http.HandleFunc("/drop-table", DropTable)

	// Inicia o servidor.
	http.ListenAndServe("localhost:9015", nil)
}
