# frozen_string_literal: true

require "rails_helper"

RSpec.describe "Accounts API", type: :request do
  let(:user) { User.create!(username: "testuser", name: "Test User", password: "Password123!") }
  let(:jwt_token) do
    post "/api/login", params: { username: user.username, password: "Password123!" }
    response.parsed_body["token"]
  end
  let(:auth_headers) { { "Authorization" => "Bearer #{jwt_token}" } }

  describe "JWT authentication" do
    it "rejects unauthenticated requests" do
      get "/accounts"
      expect(response).to have_http_status(:unauthorized)
    end

    it "accepts authenticated requests" do
      get "/accounts", headers: auth_headers
      expect(response).to have_http_status(:ok)
    end
  end

  describe "Account creation and main flag logic" do
    it "sets main to true for the first account" do
      post "/accounts", params:  { account: { name: "Main", balance: 1000, account_type: "balace" } },
                        headers: auth_headers
      expect(response).to have_http_status(:created)
      expect(response.parsed_body["main"]).to be(true)
    end

    it "does not set main to true for subsequent accounts" do
      post "/accounts", params:  { account: { name: "Main", balance: 1000, account_type: "balace" } },
                        headers: auth_headers
      post "/accounts", params:  { account: { name: "Savings", balance: 500, account_type: "savings" } },
                        headers: auth_headers
      expect(response).to have_http_status(:created)
      expect(response.parsed_body["main"]).to be(false)
    end

    it "prevents more than one main account per user" do
      post "/accounts", params:  { account: { name: "Main", balance: 1000, account_type: "balace", main: true } },
                        headers: auth_headers
      post "/accounts", params:  { account: { name: "Savings", balance: 500, account_type: "savings", main: true } },
                        headers: auth_headers
      expect(response).to have_http_status(:created)
      expect(response.parsed_body["main"]).to be(false)
    end
  end

  describe "GET /accounts/main" do
    it "returns main account info, transactions, and budgets" do
      post "/accounts", params:  { account: { name: "Main", balance: 1000, account_type: "balace" } },
                        headers: auth_headers
      get "/accounts/main", headers: auth_headers
      expect(response).to have_http_status(:ok)
      json = response.parsed_body
      expect(json["account"]["main"]).to be(true)
      expect(json["recent_transactions"]).to be_a(Array)
      expect(json["budgets"]).to be_a(Array)
    end

    it "returns 404 if no main account exists" do
      get "/accounts/main", headers: auth_headers
      expect(response).to have_http_status(:not_found)
      expect(response.parsed_body["error"]).to match(/No main account found/)
    end
  end
end
