# frozen_string_literal: true

class Account < ApplicationRecord
  belongs_to :user

  enum :type, { balace: 0, savings: 1 }
end
