# frozen_string_literal: true

FactoryBot.define do
  factory :account do
    name { "MyString" }
    balance { 1.5 }
    account_type { 1 }
    user { nil }
    main { false }
  end
end
