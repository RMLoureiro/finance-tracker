# frozen_string_literal: true

class User < ApplicationRecord
  validates :username, presence: true, uniqueness: true
  validates :name, presence: true
  validates :password, presence: true, confirmation: true, on: :create

  has_secure_password
end
