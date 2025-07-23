# frozen_string_literal: true

module Api
  class SessionsController < ApplicationController
    skip_before_action :require_login, only: %i[create]

    def create
      user = User.find_by(username: params[:username])
      if user&.authenticate(params[:password])
        token = JsonWebToken.encode(user_id: user.id)
        render json: { token: token, user: { id: user.id, username: user.username, name: user.name } }, status: :ok
      else
        render json: { error: "Invalid username or password" }, status: :unauthorized
      end
    end

    def destroy
      # No-op for JWT; frontend just deletes the token
      head :no_content
    end
  end
end
