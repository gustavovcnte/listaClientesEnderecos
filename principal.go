package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"net/http"
)

const (
	user      = "postgres"
	password  = "postgres"
	baseDados = "a3db"
	host      = "localhost"
	port      = "5432"
)

type Cliente struct {
	Id           int       `json:"id,omitempty"`
	PrimeiroNome string    `json:"nome,omitempty"`
	Sobrenome    string    `json:"sobrenome,omitempty"`
	Endereco     *Endereco `json:"endereco,omitempty"`
}

type Endereco struct {
	Id         int    `json:"endereco_id,omitempty"`
	Logradouro string `json:"logradouro,omitempty"`
	Cep        int    `json:"cep,omitempty"`
	Bairro     string `json:"bairro,omitempty"`
	Cidade     string `json:"cidade,omitempty"`
	UF         string `json:"uf,omitempty"`
}

// a3 Consulta por Cidade
type ConsultaCidade struct {
	Cidade   string           `json:"cidade,omitempty"`
	UF       string           `json:"uf,omitempty"`
	Clientes *ConsultaCliente `json:"clientes,omitempty"`
}

type ConsultaCliente struct {
	Id   int    `json:"id,omitempty"`
	Nome string `json:"nome,omitempty"`
}

// Conectar no Banco de Dados
func conectaNoBancoDeDados() *sql.DB {
	conexao := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", user, password, baseDados, host, port)
	db, err := sql.Open("postgres", conexao)
	mensagemErro(err)
	return db
}

// Mensagem de erro
func mensagemErro(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// A3 Consultar um Cliente por Cidade / ex. localhost:9090/cliente/?cidade=Palhoça

func ConsultarCidade(w http.ResponseWriter, r *http.Request) {
	cidadeURL := r.URL.Query().Get("cidade")
	db := conectaNoBancoDeDados()
	selecionarCidade, err := db.Query("select c.id, c.primeiro_nome, e.cidade, e.uf from tb_cliente c left join tb_endereco e on c.endereco_id=e.id where e.cidade=$1", cidadeURL)
	mensagemErro(err)
	var consultarc []ConsultaCidade
	var (
		id               int
		nome, cidade, uf string
	)
	consultar := ConsultaCidade{}
	consultar.Cidade = cidade
	consultar.UF = uf

	for selecionarCidade.Next() {
		err := selecionarCidade.Scan(&id, &nome, &cidade, &uf)
		mensagemErro(err)
		consultaCliente := ConsultaCliente{id, nome}
		consultar.Clientes = &consultaCliente
		consultarc = append(consultarc, consultar)
	}
	defer db.Close()
	w.Header().Set("Content-Type", "application/json")
	ccidade := ConsultaCidade{}
	ccidade.Cidade = cidade
	ccidade.UF = uf
	json.NewEncoder(w).Encode(ccidade)
	json.NewEncoder(w).Encode(consultarc)
}

// Listar os Clientes
func ListarClientes(w http.ResponseWriter, r *http.Request) {
	db := conectaNoBancoDeDados()
	listaClientes, err := db.Query("select c.id, c.primeiro_nome, c.sobrenome, c.endereco_id, e.logradouro, e.bairro, e.cep, e.cidade, e.uf from tb_cliente c left join tb_endereco e on c.endereco_id=e.id")
	mensagemErro(err)
	var clientes []Cliente
	for listaClientes.Next() {
		var (
			id, endereco_id, cep                            int
			nome, sobrenome, logradouro, cidade, bairro, uf string
		)
		err := listaClientes.Scan(&id, &nome, &sobrenome, &endereco_id, &logradouro, &bairro, &cep, &cidade, &uf)
		mensagemErro(err)
		cliente := Cliente{}
		cliente.Id = id
		cliente.PrimeiroNome = nome
		cliente.Sobrenome = sobrenome
		endereco := Endereco{endereco_id, logradouro, cep, bairro, cidade, uf}
		cliente.Endereco = &endereco
		clientes = append(clientes, cliente)

	}
	defer db.Close()
	fmt.Println(clientes)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientes)
}

// Inserir um novo cliente
func InserirUmCliente(w http.ResponseWriter, r *http.Request) {
	var cliente Cliente
	json.NewDecoder(r.Body).Decode(&cliente)
	db := conectaNoBancoDeDados()
	queryInsert, err := db.Prepare("insert into tb_cliente(id, primeiro_nome, sobrenome, endereco_id) values (nextval('my_seq_cliente'), $1, $2, $3)")
	mensagemErro(err)
	queryInsert.Exec(cliente.PrimeiroNome, cliente.Sobrenome, cliente.Endereco.Id)
	defer db.Close()
}

// Alterar dados de um cliente
func AlterarUmCliente(w http.ResponseWriter, r *http.Request) {
	parametros := mux.Vars(r)
	var clienteNovo Cliente
	err := json.NewDecoder(r.Body).Decode(&clienteNovo)
	mensagemErro(err)
	db := conectaNoBancoDeDados()
	alterarCliente, err := db.Prepare("update tb_cliente set primeiro_nome=$1, sobrenome=$2, endereco_id=$3 where id=$4")
	mensagemErro(err)
	alterarCliente.Exec(clienteNovo.PrimeiroNome, clienteNovo.Sobrenome, clienteNovo.Endereco.Id, parametros["ID"])
	defer db.Close()
}

// Deletar um cliente
func DeletarUmCliente(w http.ResponseWriter, r *http.Request) {
	parametros := mux.Vars(r)
	db := conectaNoBancoDeDados()
	deletarCliente, err := db.Prepare("delete from tb_cliente where id=$1")
	mensagemErro(err)
	deletarCliente.Exec(parametros["ID"])
	defer db.Close()
}

// Consultar por nome do cliente
func ConsultarPorNomeSobrenome(w http.ResponseWriter, r *http.Request) {
	nomeURL := r.URL.Query().Get("nome")
	sobrenomeURL := r.URL.Query().Get("sobrenome")
	db := conectaNoBancoDeDados()
	consultaNomeSobrenome, err := db.Query("select c.id, c.primeiro_nome, c.sobrenome, c.endereco_id, e.logradouro, e.bairro, e.cep, e.cidade, e.uf from tb_cliente c left join tb_endereco e on c.endereco_id=e.id where c.primeiro_nome=$1 and c.sobrenome=$2", nomeURL, sobrenomeURL)
	mensagemErro(err)
	var clientes []Cliente
	for consultaNomeSobrenome.Next() {
		var (
			id, endereco_id, cep                            int
			nome, sobrenome, logradouro, cidade, bairro, uf string
		)
		err := consultaNomeSobrenome.Scan(&id, &nome, &sobrenome, &endereco_id, &logradouro, &bairro, &cep, &cidade, &uf)
		mensagemErro(err)
		cliente := Cliente{}
		cliente.Id = id
		cliente.PrimeiroNome = nome
		cliente.Sobrenome = sobrenome
		endereco := Endereco{endereco_id, logradouro, cep, bairro, cidade, uf}
		cliente.Endereco = &endereco
		clientes = append(clientes, cliente)

	}
	defer db.Close()
	fmt.Println(clientes)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientes)
}

// Consultar por um ID do cliente
func ConsultaPorUmID(w http.ResponseWriter, r *http.Request) {
	parametros := mux.Vars(r)
	idCliente := parametros["ID"]
	db := conectaNoBancoDeDados()
	selecionarUmCliente, err := db.Query("select c.id, c.primeiro_nome, c.sobrenome, c.endereco_id, e.logradouro, e.bairro, e.cep, e.cidade, e.uf from tb_cliente c left join tb_endereco e on c.endereco_id=e.id where c.id=$1", idCliente)
	mensagemErro(err)
	cliente := Cliente{}
	for selecionarUmCliente.Next() {
		var (
			id, endereco_id, cep                            int
			nome, sobrenome, logradouro, cidade, bairro, uf string
		)
		err := selecionarUmCliente.Scan(&id, &nome, &sobrenome, &endereco_id, &logradouro, &bairro, &cep, &cidade, &uf)
		mensagemErro(err)
		cliente.Id = id
		cliente.PrimeiroNome = nome
		cliente.Sobrenome = sobrenome
		endereco := Endereco{endereco_id, logradouro, cep, bairro, cidade, uf}
		cliente.Endereco = &endereco
	}
	defer db.Close()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cliente)

}

// Main
func main() {

	rotas := mux.NewRouter()
	rotas.HandleFunc("/clientes", ListarClientes).Methods("GET")
	rotas.HandleFunc("/cliente/", ConsultarCidade).Methods("GET")
	rotas.HandleFunc("/cliente/{ID}", ConsultaPorUmID).Methods("GET")
	rotas.HandleFunc("/cliente/nome/", ConsultarPorNomeSobrenome).Methods("GET")
	rotas.HandleFunc("/cliente", InserirUmCliente).Methods("POST")
	rotas.HandleFunc("cliente/{ID}", AlterarUmCliente).Methods("PUT")
	rotas.HandleFunc("/cliente/{ID}", DeletarUmCliente).Methods("DELETE")
	http.ListenAndServe(":9090", rotas)
}
