# frozen_string_literal: true

class Account < ApplicationRecord
  belongs_to :user
  has_many :budgets, dependent: :destroy
  has_many :transactions, dependent: :destroy

  enum :account_type, { balace: 0, savings: 1 }
  validates :main, inclusion: { in: [true, false] }
  validate :only_one_main_account_per_user
  before_validation :set_main_if_first_account, on: :create

  private

  def set_main_if_first_account
    return unless user_id && Account.where(user_id: user_id, main: true).none?

    self.main = true
  end

  def only_one_main_account_per_user
    return unless main && user_id && Account.where(user_id: user_id, main: true).where.not(id: id).exists?

    errors.add(:main, "can only be set for one account per user")
  end
end
