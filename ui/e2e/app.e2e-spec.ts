import { OmckleUiPage } from './app.po';

describe('omckle-ui App', () => {
  let page: OmckleUiPage;

  beforeEach(() => {
    page = new OmckleUiPage();
  });

  it('should display message saying app works', () => {
    page.navigateTo();
    expect(page.getParagraphText()).toEqual('app works!');
  });
});
