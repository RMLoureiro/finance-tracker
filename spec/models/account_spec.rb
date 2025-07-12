# frozen_string_literal: true

require "rails_helper"

RSpec.describe Account do
  describe "structure and validations" do
    it "has all columns" do
      expect(subject).to have_db_column(:name)
      expect(subject).to have_db_column(:balance)
      expect(subject).to have_db_column(:type)
    end

    it "has relations" do
      expect(subject).to belong_to(:user)
    end
  end
end
