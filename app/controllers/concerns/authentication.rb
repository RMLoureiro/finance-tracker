# frozen_string_literal: true

module Authentication
  extend ActiveSupport::Concern

  included do
    # No CSRF needed for JWT
    helper_method :current_user
  end

  def current_user
    return @current_user if defined?(@current_user)

    header = request.headers["Authorization"]
    token = header.split.last if header
    decoded = JsonWebToken.decode(token)
    @current_user = decoded ? User.find_by(id: decoded[:user_id]) : nil
  end

  def require_login
    return if current_user

    render json: { error: "Unauthorized" }, status: :unauthorized
  end
end
