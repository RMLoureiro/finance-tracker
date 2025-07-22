# frozen_string_literal: true

module Api
  class ApplicationController < ::ApplicationController
    include Api::Authentication
    before_action :require_login
  end
end
