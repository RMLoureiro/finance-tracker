# frozen_string_literal: true

module Api
  module Authentication
    extend ActiveSupport::Concern

    included do
      protect_from_forgery with: :null_session
      helper_method :current_user
    end

    def current_user
      @current_user ||= User.find_by(id: session[:user_id]) if session[:user_id]
    end

    def require_login
      return if current_user

      render json: { error: "Unauthorized" }, status: :unauthorized
    end
  end
end
