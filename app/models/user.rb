# frozen_string_literal: true

class User < ApplicationRecord
  validates :username, uniqueness: true

  has_secure_password
end
