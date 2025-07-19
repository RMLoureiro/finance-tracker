# frozen_string_literal: true

require "rails_helper"

RSpec.describe Budget do
  describe "structure and validations" do
    it "has all columns" do
      expect(subject).to have_db_column(:name)
      expect(subject).to have_db_column(:amount)
      expect(subject).to have_db_column(:start_date)
      expect(subject).to have_db_column(:active)
    end

    it "has relations" do
      expect(subject).to belong_to(:account)
    end
  end
end
