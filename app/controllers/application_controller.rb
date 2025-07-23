# frozen_string_literal: true

class ApplicationController < ActionController::Base
  include Authentication
  before_action :set_default_format
  # Only allow modern browsers supporting webp images, web push, badges, import maps, CSS nesting, and CSS :has.
  allow_browser versions: :modern

  private

  def set_default_format
    request.format = :json
  end
end
