package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("segredo-muito-seguro")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("segredo-muito-seguro", hash) {
		t.Error("password correta rejeitada")
	}
	if VerifyPassword("segredo-errado", hash) {
		t.Error("password errada aceite")
	}
	if VerifyPassword("segredo-muito-seguro", "lixo$que$nao$e$um$hash") {
		t.Error("hash inválido aceite")
	}
}

func TestHashesAreSalted(t *testing.T) {
	a, _ := HashPassword("mesma-password")
	b, _ := HashPassword("mesma-password")
	if a == b {
		t.Error("dois hashes da mesma password não deviam ser iguais (salt)")
	}
}
