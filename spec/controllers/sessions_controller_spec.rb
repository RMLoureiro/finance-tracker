# frozen_string_literal: true

require "rails_helper"

RSpec.describe "Sessions API", type: :request do
  let!(:user) { User.create!(username: "testuser", name: "Test User", password: "Password123!") }

  describe "POST /api/login" do
    it "returns a JWT token for valid credentials" do
      post "/api/login", params: { username: user.username, password: "Password123!" }
      expect(response).to have_http_status(:ok)
      json = response.parsed_body
      expect(json["token"]).to be_present
      expect(json["user"]["username"]).to eq(user.username)
    end

    it "returns 401 for invalid credentials" do
      post "/api/login", params: { username: user.username, password: "wrongpassword" }
      expect(response).to have_http_status(:unauthorized)
      json = response.parsed_body
      expect(json["error"]).to match(/Invalid username or password/)
    end
  end
end
