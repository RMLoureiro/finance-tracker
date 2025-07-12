# frozen_string_literal: true

FactoryBot.define do
  factory :account do
    name { "MyString" }
    balance { 1.5 }
    type { 1 }
    user { nil }
  end
end
