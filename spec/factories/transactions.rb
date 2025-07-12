# frozen_string_literal: true

FactoryBot.define do
  factory :transaction do
    ammount { 1.5 }
    user { nil }
    transaction_types { nil }
  end
end
