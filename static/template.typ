#import "@preview/codetastic:0.2.2": qrcode

#let GENERAL_STROKE = 0.8pt
#let HEADER_SIZE = 21pt
#let HEADER_SPACING = 3.5mm
#let DEBUG = sys.inputs.at("debug", default: false)
#let DOSSIER_FILE = sys.inputs.at("input", default: "dossier.json")

// ========================================
// DATA
// ========================================

#let dossier = json(if DEBUG { "test-dossier.json" } else { DOSSIER_FILE })

// ========================================
// METHODS
// ========================================

#let render_transaction_table(attachment) = table(

)

#let render_header(attachment, is_fist_page) = {
  set par(spacing: 0mm)
  grid(
    columns: (1fr, 17mm),
    stroke: GENERAL_STROKE,
    grid.cell(
      inset: 2.5mm,
      [
        #set par(spacing: HEADER_SPACING)
        #set text(size: 13pt)
        #if is_fist_page [
          = #attachment.Path
        ] else [
          #set text(size: HEADER_SIZE)
          *#sym.arrow #attachment.Path*
        ]

        #attachment.FileName -- #attachment.Transactions.first().Ident
      ],
    ),
    grid.cell(
      align: horizon + center,
      qrcode(attachment.Path, width: 13.5mm, quiet-zone: 0),
    ),
  )
  if is_fist_page {
    render_transaction_table(attachment)
  }
}

#let render_content(attachment) = {
  image(attachment.FileUUID)
}

#let render_footer(current_page, total_pages) = {
  set par(leading: 0.5em)
  grid(
    columns: (1fr, 1fr, 1fr),
    align: (left, center, right),
    inset: 1.5mm,
    [
      #dossier.CompanyName\
      #dossier.Street\
      #dossier.ZIPCode #dossier.Place\
    ],
    [
      *#current_page / #total_pages -- Page TODO*\
      TODO -- TODO (Transaction date range)
    ],
    [
      File: #dossier.AccountingFileName\
      Accounting data as of: TODO\
      Report was created on: TODO
    ]
  )
}


#let render_attachment(attachment) = {
  for page in range(attachment.PageCount) {
    grid(
      columns: 1fr,
      rows: (auto, 1fr, auto),
      stroke: GENERAL_STROKE,
      render_header(attachment, page == 0),
      render_content(attachment),
      render_footer(page + 1, attachment.PageCount),
    )
  }
}

// ========================================
// LAYOUT
// ========================================

#set page(
  margin: 10mm,
)

#show heading.where(level: 1): set block(below: HEADER_SPACING)
#show heading.where(level: 1): set text(size: HEADER_SIZE)

// ========================================
// CONTENT
// ========================================

#for attachment in dossier.JournalEntries {
  render_attachment(attachment)
}
