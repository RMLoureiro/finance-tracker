# frozen_string_literal: true

class UsersController < ApplicationController
  include Authentication
  before_action :require_login
  skip_before_action :require_login, only: [:create]

  def show
    render json: { id: current_user.id, username: current_user.username, name: current_user.name }
  end

  def create
    user = User.new(user_params)
    if user.save
      render json: { id: user.id, username: user.username, name: user.name }, status: :created
    else
      render json: { errors: user.errors.full_messages }, status: :unprocessable_entity
    end
  end

  private

  def user_params
    params.expect(user: %i[username name password])
  end
end
