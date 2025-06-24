# frozen_string_literal: true

FactoryBot.define do
  factory :user do
    username { "TEST" }
    name { "TEST NAME" }
    password_digest { "password" }
  end
end
