# frozen_string_literal: true

module Api
  class SessionsController < ApplicationController
    protect_from_forgery with: :null_session
    skip_before_action :require_login, only: %i[create destroy]

    def create
      user = User.find_by(username: params[:username])
      if user&.authenticate(params[:password])
        session[:user_id] = user.id
        render json: { id: user.id, username: user.username, name: user.name }, status: :ok
      else
        render json: { error: "Invalid username or password" }, status: :unauthorized
      end
    end

    def destroy
      session[:user_id] = nil
      head :no_content
    end
  end
end
