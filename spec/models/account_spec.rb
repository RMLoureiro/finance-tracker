# frozen_string_literal: true

require "rails_helper"

RSpec.describe Account do
  describe "structure and validations" do
    it "has all columns" do
      expect(subject).to have_db_column(:name)
      expect(subject).to have_db_column(:balance)
      expect(subject).to have_db_column(:account_type)
    end

    it "has relations" do
      expect(subject).to belong_to(:user)
    end
  end

  describe "main flag logic" do
    let(:user) { User.create!(username: "mainflaguser", name: "Main Flag User", password: "Password123!") }

    it "sets main to true for the first account" do
      account = user.accounts.create!(name: "Main", balance: 1000, account_type: :balace)
      expect(account.main).to be(true)
    end

    it "sets main to false for subsequent accounts" do
      user.accounts.create!(name: "Main", balance: 1000, account_type: :balace)
      account2 = user.accounts.create!(name: "Savings", balance: 500, account_type: :savings)
      expect(account2.main).to be(false)
    end

    it "prevents more than one main account per user via validation" do
      user.accounts.create!(name: "Main", balance: 1000, account_type: :balace)
      account2 = user.accounts.build(name: "Savings", balance: 500, account_type: :savings, main: true)
      expect(account2).not_to be_valid
      expect(account2.errors[:main].join).to match(/can only be set for one account per user/)
    end
  end
end
