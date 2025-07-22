# frozen_string_literal: true

require "rails_helper"

RSpec.describe UsersController do
  describe "#show" do
    let(:user) { User.create!(username: "testuser", name: "Test User", password: "Password123!") }

    before do
      allow(controller).to receive(:current_user).and_return(user)
    end

    it "renders the current user as JSON" do
      get :show
      expect(response).to have_http_status(:ok)
      json = response.parsed_body
      expect(json).to include("id" => user.id, "username" => user.username, "name" => user.name)
    end
  end

  describe "#create" do
    context "with valid params" do
      let(:user_params) { { username: "newuser", name: "New User", password: "Password123!" } }

      it "creates a new user and returns JSON" do
        expect do
          post :create, params: { user: user_params }
        end.to change(User, :count).by(1)
        expect(response).to have_http_status(:created)
        json = response.parsed_body
        expect(json).to include("id", "username" => "newuser", "name" => "New User")
      end
    end

    context "with invalid params" do
      let(:invalid_params) { { username: "", name: "", password: "" } }

      it "does not create a user and returns errors" do
        expect do
          post :create, params: { user: invalid_params }
        end.not_to change(User, :count)
        expect(response).to have_http_status(:unprocessable_entity)
        json = response.parsed_body
        expect(json["errors"]).to include(
          "Username can't be blank",
          "Name can't be blank",
          "Password can't be blank"
        )
      end

      it "permits only username, name, and password" do
        params = ActionController::Parameters.new(user: { username: "a", name: "b", password: "c", admin: true })
        controller.params = params
        permitted = controller.send(:user_params)
        expect(permitted.keys).to contain_exactly("username", "name", "password")
      end
    end
  end
end
